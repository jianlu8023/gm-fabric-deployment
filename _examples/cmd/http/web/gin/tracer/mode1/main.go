package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"_examples/cmd/http/web/gin/tracer/mode1/tracer"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func Register(e *gin.Engine) {
	e.GET("/ping", Ping)
}

func Ping(c *gin.Context) {
	psc, _ := c.Get("ctx")
	ctx := psc.(context.Context)
	doPing1(ctx)
	doPing2(ctx)
	c.JSON(200, gin.H{"message": "pong"})
}
func doPing1(ctx context.Context) {
	span, _ := opentracing.StartSpanFromContext(ctx, "doPing1")
	defer span.Finish()
	time.Sleep(time.Second)
	fmt.Println("pong")
}
func doPing2(ctx context.Context) {
	span, _ := opentracing.StartSpanFromContext(ctx, "doPing2")
	defer span.Finish()
	time.Sleep(time.Second)
	fmt.Println("pong")
}

func Jaeger() gin.HandlerFunc {
	return func(c *gin.Context) {
		var parentSpan opentracing.Span
		newTracer, closer := tracer.NewTracer("gin-demo")
		defer closer.Close()
		// 直接从 c.Request.Header 中提取 span,如果没有就新建一个
		spCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))
		if err != nil {
			parentSpan = newTracer.StartSpan(c.Request.URL.Path)
			defer parentSpan.Finish()
		} else {
			parentSpan = opentracing.StartSpan(
				c.Request.URL.Path,
				opentracing.ChildOf(spCtx),
				opentracing.Tag{Key: string(ext.Component), Value: "HTTP"},
				ext.SpanKindRPCServer,
			)
			defer parentSpan.Finish()
		}
		c.Set("newTracer", newTracer)
		c.Set("ctx", opentracing.ContextWithSpan(context.Background(), parentSpan))
		c.Next()
	}
}

func main() {

	e := gin.New()
	e.Use(Jaeger())
	Register(e)
	server := &http.Server{
		Addr:         ":8080",
		Handler:      e,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server listen: %s\n", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	sig := <-signalChan
	log.Println("Get Signal:", sig)
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
