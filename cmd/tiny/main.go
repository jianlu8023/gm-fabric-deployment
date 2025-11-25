package main

import (
	"fmt"
	"html/template"
	gohttp "net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/colour"
	"github.com/jianlu8023/go-tools/v2/pkg/hash/md5"
	"github.com/jianlu8023/golang-example/cmd/tiny/internal/command"
	"github.com/jianlu8023/golang-example/cmd/tiny/internal/conf"
	"github.com/jianlu8023/golang-example/cmd/tiny/internal/container"
	"github.com/jianlu8023/golang-example/cmd/tiny/internal/container/verify"
	"github.com/jianlu8023/golang-example/cmd/tiny/internal/server/handle/core"
	middleware2 "github.com/jianlu8023/golang-example/cmd/tiny/internal/server/middleware"
	"github.com/jianlu8023/golang-example/cmd/tiny/templates"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/certificate"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/http"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"go.uber.org/zap"
)

func main() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	var err error

	loggerControl := logger.NewLoggerControl(&config.LoggerConfig{
		DefaultLogLevel: "info",
		PrintFormat:     "console",
	})
	if loggerControl == nil {
		return
	}
	loggerControl.StartUp(func(err error) {
		if err != nil {
			quit <- os.Interrupt
		}
	})

	mainLogger := loggerControl.GenLogger("")
	defer func() {
		_ = loggerControl.Shutdown()

		if err != nil {
			mainLogger.Errorf("启动失败: %v", err)
		}
	}()

	certPath, err := conf.GenConfigPath()
	if err != nil {
		mainLogger.Errorf("获取配置文件路径失败: %v", err)
		return
	}

	certificateControl := certificate.NewCertificateControl(&config.CertificateConfig{
		Enabled:     true,
		CertPath:    certPath,
		DefaultAlgo: "RSA",
		RootSubject: "/CN=Tiny Root CA",
	}, loggerControl)

	if certificateControl == nil {
		return
	}

	certificateControl.StartUp(func(err error) {
		if err != nil {
			quit <- os.Interrupt
		}
	})

	defer func() {
		_ = certificateControl.Shutdown()
	}()

	if err = conf.LoadConfig(mainLogger, certificateControl); err != nil {
		mainLogger.Errorf("加载配置文件失败: %v", err)
		return
	}

	if err = command.CLI().Run(os.Args); err != nil {
		mainLogger.Errorf("解析启用命令失败: %v", err)
		return
	}

	if err = container.NewHandleChain().
		AddToHead(verify.NewPortVerifier(conf.Config.Port)).
		AddToHead(verify.NewPathVerifier(conf.Config.RootPath)).
		Iterator(); err != nil {
		mainLogger.Errorf("链式检查必要配置失败: %v", err)
		return
	}

	httpControl, err := http.NewWebServerControl(&config.HttpServerConfig{
		Enabled:        true,
		Address:        ":" + strconv.Itoa(conf.Config.Port),
		RunMode:        "release",
		TlsEnabled:     conf.Config.TlsEnabled,
		TlsGM:          false,
		TlsKeyFile:     conf.Config.TlsKeyPath,
		TlsCertFile:    conf.Config.TlsCertPath,
		TlsRCACertFile: conf.Config.TlsRCACertPath,
		Http2Enabled:   true,
		Pprof:          false,
		UploadDir:      "",
	},
		loggerControl,
	)
	if err != nil {
		mainLogger.Errorf("启动http服务失败: %v", err)
		return
	}
	if httpControl == nil {
		return
	}
	defer func() {
		_ = httpControl.Shutdown()
	}()

	t, err := template.ParseFS(templates.FS, "*.tpl")
	if err != nil {
		mainLogger.Errorf("template parse error: %v", err)
		quit <- os.Interrupt
	}
	httpControl.RegisterHtmlTemplate(t)

	httpControl.RegisterRouter([]commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:   "index",
			Uri:    "/",
			Method: gohttp.MethodGet,
			HandlerFunc: func(ctx *gin.Context) {
				ctx.Redirect(gohttp.StatusPermanentRedirect, "/file/")
			},
			EnableJWtVerify: false,
			Enabled:         true,
			Desc:            "index",
		},
		&commonhttp.MyRouter{
			Name:   "login-get",
			Uri:    "login",
			Method: gohttp.MethodGet,
			HandlerFunc: func(ctx *gin.Context) {
				ctx.HTML(gohttp.StatusOK, "login.tpl", nil)
			},
			Enabled:         true,
			EnableJWtVerify: false,
			Desc:            "login-get",
		},
		&commonhttp.MyRouter{
			Name:   "login-post",
			Uri:    "login",
			Method: gohttp.MethodPost,
			HandlerFunc: func(ctx *gin.Context) {
				// 检查帐号密码
				// 通过则生成session，跳转首页
				// 不通过则返回登录页

				if md5.SumStringHex(ctx.PostForm("username")) == conf.Config.Username &&
					md5.SumStringHex(ctx.PostForm("password")) == conf.Config.Password {
					// 由于session相关代码被注释，我们设置一个简单的cookie用于验证
					ctx.SetCookie("login", conf.Config.SessionVal, 3600, "/", "", conf.Config.TlsEnabled, true)
					ctx.JSON(gohttp.StatusOK, gin.H{"code": 1, "message": "登录成功"})
					return
				} else {
					ctx.JSON(gohttp.StatusOK, gin.H{"code": 0, "message": "登录失败"})
				}
			},
			Enabled:         true,
			EnableJWtVerify: false,
			Desc:            "login-post",
		},
		&commonhttp.MyRouter{
			Name:   "ico",
			Uri:    "/favicon.ico",
			Method: gohttp.MethodGet,
			HandlerFunc: func(ctx *gin.Context) {
				file, _ := templates.FS.ReadFile("favicon.ico")
				ctx.Data(gohttp.StatusOK, "image/x-icon", file)
			},
			EnableJWtVerify: false,
			Enabled:         true,
			Desc:            "ico",
		},
	})
	httpControl.RegisterGroupedRouter(&commonhttp.MyGroupRouter{
		Group: conf.FileGroupPrefix,
		Routers: []commonhttp.RouterHandler{
			&commonhttp.MyRouter{
				Name:            "download",
				Uri:             "/*filename",
				Method:          gohttp.MethodGet,
				HandlerFunc:     core.Downloader,
				Enabled:         true,
				EnableJWtVerify: false,
				Desc:            "download",
			},
			&commonhttp.MyRouter{
				Name:            "upload",
				Uri:             "/upload",
				Method:          gohttp.MethodPost,
				EnableJWtVerify: false,
				Enabled:         true,
				HandlerFunc: func(ctx *gin.Context) {
					f, err := ctx.FormFile("upload_file")
					if err != nil {
						commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "文件上传失败!")
						return
					}

					currPath := ctx.PostForm("path")
					err = ctx.SaveUploadedFile(f, filepath.Join(conf.Config.RootPath, currPath, f.Filename))
					if err != nil {
						commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "文件保存失败!")
						return
					}
					ctx.Redirect(gohttp.StatusMovedPermanently, path.Join(conf.FileGroupPrefix, currPath))
				},
				Desc: "upload",
			},
		},
		MiddlewaresFunc: []gin.HandlerFunc{
			middleware2.CheckLevel,
			middleware2.CheckLogin,
		},
	})

	httpControl.StartUp(func(err error) {
		if err != nil {
			quit <- os.Interrupt
		}
	})

	printInfo(mainLogger)

	<-quit
	mainLogger.Infof("received shutdown signal...")
	time.Sleep(5 * time.Second)
}

// printInfo 会在程序启动后打印本机 IP、共享目录、是否允许上传的信息
func printInfo(logger *zap.SugaredLogger) {
	// Print IP information
	if conf.Config.IP != "" {
		protocol := "http"
		if conf.Config.TlsEnabled {
			protocol = "https"
		}
		logger.Infof("Run on   [ %s ]", colour.Blue(fmt.Sprintf("%s://%s:%d", protocol, conf.Config.IP, conf.Config.Port)))
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
	logger.Infof("Need Login: [ %s ]", status)

	// Print TLS status
	status = colour.Red(fmt.Sprintf("%t", conf.Config.TlsEnabled))
	if conf.Config.TlsEnabled {
		status = colour.Green(fmt.Sprintf("%t", conf.Config.TlsEnabled))
	}
	logger.Infof("TLS Enabled: [ %s ]", status)
}
