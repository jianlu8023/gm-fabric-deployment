package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"

	"_examples/pkg/tracing"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	traceapi "go.opentelemetry.io/otel/trace"
)

// var tracer = otel.Tracer("gin-server")
//
// func main() {
//
//		tp, err := initTracer()
//		if err != nil {
//			log.Fatal(err)
//		}
//		defer func() {
//			if err := tp.Shutdown(context.Background()); err != nil {
//				log.Printf("Error shutting down tracer provider: %v", err)
//			}
//		}()
//		r := gin.New()
//		r.Use(otelgin.Middleware("my-server"))
//		tmplName := "user"
//		tmplStr := "user {{ .name }} (id {{ .id }})\n"
//		tmpl := template.Must(template.New(tmplName).Parse(tmplStr))
//		r.SetHTMLTemplate(tmpl)
//		r.GET("/users/:id", func(c *gin.Context) {
//			id := c.Param("id")
//			name := getUser(c, id)
//			otelgin.HTML(c, http.StatusOK, tmplName, gin.H{
//				"name": name,
//				"id":   id,
//			})
//		})
//		_ = r.Run(":8080")
//	}
//
//	func initTracer() (*sdktrace.TracerProvider, error) {
//		exporter, err := stdout.New(stdout.WithPrettyPrint())
//		if err != nil {
//			return nil, err
//		}
//		tp := sdktrace.NewTracerProvider(
//			sdktrace.WithSampler(sdktrace.AlwaysSample()),
//			sdktrace.WithBatcher(exporter),
//		)
//		otel.SetTracerProvider(tp)
//		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
//		return tp, nil
//	}
func getUser(c *gin.Context, id string) string {
	// Pass the built-in `context.Context` object from http.Request to OpenTelemetry APIs
	// where required. It is available from gin.Context.Request.Context()
	_, span := tracing.Span(c.Request.Context(), "gin-web", "getUser", traceapi.WithAttributes(attribute.String("id", id)))
	defer span.End()
	span.AddEvent("getUser", traceapi.WithAttributes(attribute.String("userId", id)))
	if id == "123" {
		return "otelgin tester"
	}
	return "unknown"
}

func main() {

	os.Setenv("OTEL_EXPORTER_FILE_PATH", "./traces.json")
	os.Setenv("OTEL_EXPORTER_OTLP_INSECURE", "true")
	os.Setenv("OTEL_TRACES_EXPORTER", "otlp,file")
	// os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://192.168.58.110:4318")
	os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://192.168.58.110:4317")
	os.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
	os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "grpc")
	ctx := context.Background()

	tp, err := tracing.NewTracerProvider(ctx, "go-example-gin-web")
	if err != nil {
		log.Fatalf("failed to initialize tracer provider: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown tracer provider: %v", err)
		}
	}()

	r := gin.New()
	r.Use(otelgin.Middleware(""))
	tmplName := "user"
	tmplStr := "user {{ .name }} (id {{ .id }})\n"
	tmpl := template.Must(template.New(tmplName).Parse(tmplStr))
	r.SetHTMLTemplate(tmpl)
	r.GET("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		name := getUser(c, id)
		otelgin.HTML(c, http.StatusOK, tmplName, gin.H{
			"name": name,
			"id":   id,
		})
	})
	_ = r.Run(":8080")
}
