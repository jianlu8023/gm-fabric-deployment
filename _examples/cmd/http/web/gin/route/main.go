package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"_examples/internal/logger"
	"github.com/gin-gonic/gin"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc gin.HandlerFunc
	Enabled     bool
}

func Routes() []Route {

	routeGroups := map[string][]Route{
		"base": {
			{
				Name:    "status",
				Method:  http.MethodGet,
				Pattern: "/",
				HandlerFunc: func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{
						"message": "ok",
					})
				},
				Enabled: true,
			},

			{
				Name:    "version",
				Method:  http.MethodGet,
				Pattern: "/version",
				HandlerFunc: func(ctx *gin.Context) {

				},
				Enabled: false,
			},
			{
				Name:    "route",
				Method:  http.MethodGet,
				Pattern: "/route",
				HandlerFunc: func(ctx *gin.Context) {
					routes := Routes()
					type r struct {
						Method  string
						Pattern string
						Enabled bool
					}
					var rs []r
					for _, route := range routes {
						rs = append(rs, r{
							Method:  route.Method,
							Pattern: route.Pattern,
							Enabled: route.Enabled,
						})
					}
					ctx.JSON(http.StatusOK, rs)

				},
				Enabled: true,
			},
		},
	}

	sum := 0
	for _, routeGroup := range routeGroups {
		sum += len(routeGroup)
	}

	routes := make([]Route, 0, sum)
	for _, routeGroup := range routeGroups {
		routes = append(routes, routeGroup...)
	}

	return routes
}

func main() {
	router := gin.Default()

	for _, route := range Routes() {
		if route.Enabled {
			switch route.Method {
			case http.MethodGet:
				router.GET(fmt.Sprintf("/%s%s", "demo", route.Pattern), route.HandlerFunc)
			case http.MethodPost:
				router.POST(fmt.Sprintf("/%s%s", "demo", route.Pattern), route.HandlerFunc)
			}
		}
	}

	srv := http.Server{
		Handler: router,
		Addr:    ":8080",
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.GetAppLogger().Errorf("启动gin web服务失败 %v", err)
			quit <- syscall.SIGINT
		}
	}()

	<-quit
	if err := srv.Shutdown(context.Background()); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.GetAppLogger().Errorf("关闭gin web服务失败 %v", err)
	}

}
