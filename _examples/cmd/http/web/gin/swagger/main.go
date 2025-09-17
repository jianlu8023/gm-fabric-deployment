package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	_ "_examples/cmd/http/web/gin/swagger/docs"
	"_examples/internal/logger"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Article struct {
	ID            uint32 `gorm:"primary_key" json:"id" form:"tag_id"`
	CreatedBy     string `json:"created_by" form:"create_by"`
	ModifiedBy    string `json:"modified_by" form:"modify_by"`
	CreatedOn     uint32 `json:"created_on" form:"create_on"`
	ModifiedOn    uint32 `json:"modified_on" form:"modify_on"`
	DeletedOn     uint32 `json:"deleted_on" form:"delete_on"`
	IsDel         uint8  `json:"is_del" `
	Title         string `json:"title" form:"title"`
	Desc          string `json:"desc" form:"desc"`
	Content       string `json:"content" form:"content"`
	CoverImageUrl string `json:"cover_image_url" form:"cover_image_url"`
	State         uint8  `json:"state" form:"state"`
}

func NewArticle() Article {
	return Article{}
}

var (
	articles = make(map[uint32]*Article)
)

// Get article
// @Summary 获取单个文章
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} Article "成功"
// @Failure 400 {object} string "请求错误"
// @Failure 500 {object} string "内部错误"
// @Router /api/v1/articles/{id} [get]
func (a Article) Get(c *gin.Context) {
	logger.GetAppLogger().Debugf("Get")
	param := c.Param("id")
	logger.GetAppLogger().Debugf(">>> Get Article, param: %s", param)
	articleId, err := strconv.ParseUint(param, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed",
			"code":    http.StatusBadRequest,
			"success": false,
			"data":    err.Error(),
		})
		return
	}

	article, ok := articles[uint32(articleId)]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed",
			"code":    http.StatusBadRequest,
			"success": false,
			"data":    errors.New("article not found"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"code":    http.StatusOK,
		"success": true,
		"data":    article,
	})
}

// List article
// @Summary 获取多个文章
// @Produce json
// @Param name query string false "文章名称"
// @Param tag_id query int false "标签ID"
// @Param state query int false "状态"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} Article "成功"
// @Failure 400 {object} string "请求错误"
// @Failure 500 {object} string "内部错误"
// @Router /api/v1/articles [get]
func (a Article) List(c *gin.Context) {
	logger.GetAppLogger().Debugf("List")
	name := c.Param("name")
	tagId := c.Param("tag_id")
	state := c.Param("state")
	page := c.Param("page")
	pageSize := c.Param("page_size")
	if len(name) == 0 || len(tagId) == 0 || len(state) == 0 || len(page) == 0 || len(pageSize) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed",
			"code":    http.StatusBadRequest,
			"success": false,
			"data":    errors.New("param is empty"),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"code":    http.StatusOK,
		"success": true,
		"data":    articles,
	})
	return
}

// Create article
// @Summary 创建文章
// @Produce json
// @Param tag_id body uint true "标签ID"
// @Param title body string true "文章标题"
// @Param desc body string false "文章简述"
// @Param cover_image_url body string true "封面图片地址"
// @Param content body string true "文章内容"
// @Param created_by body int true "创建者"
// @Param state body int false "状态"
// @Success 200 {object} string "成功"
// @Failure 400 {object} string "请求错误"
// @Failure 500 {object} string "内部错误"
// @Router /api/v1/articles [post]
func (a Article) Create(c *gin.Context) {
	logger.GetAppLogger().Debugf("Create")
	var article Article

	if err := c.ShouldBind(&article); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed",
			"code":    http.StatusBadRequest,
			"success": false,
			"data":    err.Error(),
		})
		return
	}
	articles[article.ID] = &article
	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"code":    http.StatusOK,
		"success": true,
		"data":    article,
	})
}

// Update article
// @Summary 更新文章
// @Produce json
// @Param tag_id body string false "标签ID"
// @Param title body string false "文章标题"
// @Param desc body string false "文章简述"
// @Param cover_image_url body string false "封面图片地址"
// @Param content body string false "文章内容"
// @Param modified_by body string true "修改者"
// @Success 200 {object} Article "成功"
// @Failure 400 {object} string "请求错误"
// @Failure 500 {object} string "内部错误"
// @Router /api/v1/articles/{id} [put]
func (a Article) Update(c *gin.Context) {
	logger.GetAppLogger().Debugf("Update")
	var article Article
	if err := c.ShouldBind(&article); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed",
			"code":    http.StatusBadRequest,
			"success": false,
			"data":    err.Error(),
		})
		return
	}

	exist, ok := articles[article.ID]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed",
			"code":    http.StatusBadRequest,
			"success": false,
			"data":    errors.New("article not exist"),
		})
		return
	}

	exist.Title = article.Title
	exist.Desc = article.Desc
	exist.Content = article.Content
	exist.CoverImageUrl = article.CoverImageUrl
	exist.State = article.State
	exist.ModifiedBy = article.ModifiedBy
	articles[article.ID] = exist
	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"code":    http.StatusOK,
		"success": true,
		"data":    exist,
	})
	return
}

// Delete article
// @Summary 删除文章
// @Produce  json
// @Param id path int true "文章ID"
// @Success 200 {string} string "成功"
// @Failure 400 {object} string "请求错误"
// @Failure 500 {object} string "内部错误"
// @Router /api/v1/articles/{id} [delete]
func (a Article) Delete(c *gin.Context) {
	logger.GetAppLogger().Debugf("Delete")
	id := c.Param("id")
	if len(id) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed",
			"code":    http.StatusBadRequest,
			"success": false,
			"data":    errors.New("id is empty"),
		})
		return
	}
	articleId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed",
			"code":    http.StatusBadRequest,
			"success": false,
			"data":    err.Error(),
		})
		return
	}

	if ar, ok := articles[uint32(articleId)]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed",
			"code":    http.StatusBadRequest,
			"success": false,
			"data":    errors.New("article not exist"),
		})
		return
	} else {
		ar.IsDel = 1
		articles[ar.ID] = ar
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"code":    http.StatusOK,
		"success": true,
		"data":    "success",
	})
	return
}

func main() {

	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Any("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "ok",
			"code":    http.StatusOK,
			"success": true,
		})
	})
	article := NewArticle()
	articles := router.Group("/api/v1/articles")
	{
		articles.GET("/:id", article.Get)
		articles.GET("/", article.List)
		articles.POST("/", article.Create)
		articles.PUT("/:id", article.Update)
		articles.DELETE("/:id", article.Delete)

	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
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
	if err := srv.Shutdown(context.Background()); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		logger.GetAppLogger().Errorf("关闭gin web服务失败 %v", err)
	}

}
