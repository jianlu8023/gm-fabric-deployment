package core

import (
	"net/http"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/cmd/tiny/handle"
	"github.com/jianlu8023/golang-example/cmd/tiny/internal/conf"
)

func Uploader(c *gin.Context) {
	f, err := c.FormFile("upload_file")
	if err != nil {
		handle.ErrorHandle(c, "文件上传失败！")
		return
	}

	currPath := c.PostForm("path")
	err = c.SaveUploadedFile(f, filepath.Join(conf.Config.RootPath, currPath, f.Filename))
	if err != nil {
		handle.ErrorHandle(c, "文件保存失败！")
		return
	}
	c.Redirect(http.StatusMovedPermanently, path.Join(conf.FileGroupPrefix, currPath))
}
