package router

import (
	"html/template"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/cmd/tiny/templates"
	"go.uber.org/zap"
)

func Setup(r *gin.Engine, logger *zap.SugaredLogger) *gin.Engine {

	t, err := template.ParseFS(templates.FS, "*.tpl")
	if err != nil {
		logger.Errorf("template parse error: %v", err)
	}
	r.SetHTMLTemplate(t)

	loadIndexRoute(r)
	loadCoreRoute(r)
	loadLoginRoute(r)
	load404Route(r)
	loadICORoute(r)
	return r
}
