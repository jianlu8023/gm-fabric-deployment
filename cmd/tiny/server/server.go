package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/colour"
	"github.com/jianlu8023/golang-example/cmd/tiny/internal/conf"
	"github.com/jianlu8023/golang-example/cmd/tiny/server/middleware"
	"github.com/jianlu8023/golang-example/cmd/tiny/server/router"
	"go.uber.org/zap"
)

// RunCore 函数负责启动 gin 实例，开始提供 HTTP 服务
func RunCore(logger *zap.SugaredLogger) {
	var (
		srv = initServer(logger)
		q   = make(chan os.Signal, 1)
	)
	signal.Notify(q, syscall.SIGINT, syscall.SIGTERM)

	{
		go run(srv, logger)
		<-q
		exit(srv, logger)
	}
}

func initServer(logger *zap.SugaredLogger) *http.Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	middleware.Setup(r)
	router.Setup(r, logger)
	s := &http.Server{
		Addr:    ":" + strconv.Itoa(conf.Config.Port),
		Handler: r,
	}
	return s
}

func run(srv *http.Server, logger *zap.SugaredLogger) {
	printInfo(logger)

	
	if conf.Config.TlsEnabled{
	
	}

	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Errorf("启用服务失败: %v", err)
	}
}

func exit(srv *http.Server, logger *zap.SugaredLogger) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("关闭服务失败: %v", err)
	}
	fmt.Println(colour.Green("\nbye"))
	os.Exit(0)
}

// printInfo 会在程序启动后打印本机 IP、共享目录、是否允许上传的信息
func printInfo(logger *zap.SugaredLogger) {
	// Print IP information
	if conf.Config.IP != "" {
		logger.Infof("Run on   [ %s ]", colour.Blue(fmt.Sprintf("http://%s:%d", conf.Config.IP, conf.Config.Port)))
	} else {
		logger.Infof("%s", colour.Yellow("Warning: [ 暂时获取不到您的IP，可以打开新的命令行窗口输入 ->  ipconfig , 查看您的IP。]"))
	}

	// Print RootPath information
	logger.Infof("Run with [ %s ]", colour.Blue(fmt.Sprintf("%s", conf.Config.RootPath)))

	// Print Max allow access level
	logger.Infof("Allow access level: [ %s ]", colour.Blue(fmt.Sprintf("%d", conf.Config.MaxLevel)))

	// Print Allow upload Status
	status := colour.Red(fmt.Sprintf("%t", conf.Config.IsAllowUpload))
	if conf.Config.IsAllowUpload {
		status = colour.Green(fmt.Sprintf("%t", conf.Config.IsAllowUpload))
	}
	logger.Infof("Allow upload: [ %s ]", status)

	// Print Secure status
	status = colour.Red(fmt.Sprintf("%t", conf.Config.IsSecure))
	if conf.Config.IsSecure {
		status = colour.Green(fmt.Sprintf("%t", conf.Config.IsSecure))
	}
	logger.Infof("Need Login: [ %s ]\n\n", status)
}
