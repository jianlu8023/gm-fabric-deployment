package router

import (
	"html/template"
	"log"
	
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/cmd/tiny/templates"
)

func Setup(r *gin.Engine) *gin.Engine {
	
	t, err := template.ParseFS(templates.FS, "*.tpl")
	if err != nil {
		log.Fatal(err.Error())
	}
	r.SetHTMLTemplate(t)
	
	loadIndexRoute(r)
	loadCoreRoute(r)
	loadLoginRoute(r)
	load404Route(r)
	loadICORoute(r)
	return r
}
