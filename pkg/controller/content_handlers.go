package controller

import (
	"fmt"
	"os"
	"time"

	"git.aetherial.dev/aeth/keiji/pkg/helpers"
	"github.com/gin-gonic/gin"
)


func (c *Controller) ServeBlogDirectory(ctx *gin.Context) {
	ctx.HTML(200, "admin", gin.H{
		"navigation": gin.H{
			"menu": c.Menu(),
			"headers": c.Headers().Elements,
		},
		"Tables": c.FormatDocTable().Tables,

	})

}


func (c *Controller) GetBlogPostEditor(ctx *gin.Context) {
	rds := helpers.NewRedisClient(helpers.RedisConf{Addr: os.Getenv("REDIS_ADDR"), Port: os.Getenv("REDIS_PORT")})
	post, exist := ctx.Params.Get("post-name")
	if !exist {
		ctx.JSON(404, map[string]string{
			"Error": "the requested file could not be found",
		})
		return
	}
	doc, err := rds.GetItem(post)
	if err != nil {
		ctx.JSON(500, map[string]string{
			"Error": err.Error(),
		})
		return
	}
	ctx.HTML(200, "blogpost_editor", gin.H{
		"navigation": gin.H{
			"menu": c.Menu(),
			"headers": c.Headers().Elements,
		},
		"HttpMethod": "patch",
		"Ident": doc.Ident,
		"Topics": helpers.Topics,
		"DefaultTopic": doc.Category,
		"Created": doc.Created,
		"Body": doc.Body,

	})
}

func (c *Controller) UpdateBlogPost(ctx *gin.Context) {
	var doc helpers.Document

	err := ctx.ShouldBind(&doc); if err != nil {
		ctx.HTML(500, "upload_status", gin.H{"UpdateMessage": "Update Failed!", "Color": "red"})
		return
	}
	rds := helpers.NewRedisClient(helpers.RedisConf{Addr: os.Getenv("REDIS_ADDR"), Port: os.Getenv("REDIS_PORT")})
	err = rds.UpdatePost(doc.Ident, doc); if err != nil {
		ctx.HTML(400, "upload_status", gin.H{"UpdateMessage": "Update Failed!", "Color": "red"})
		return
	}
	ctx.HTML(200, "upload_status", gin.H{"UpdateMessage": "Update Successful!", "Color": "green"})

}



func (c *Controller) ServeNewBlogPage(ctx *gin.Context) {

	ctx.HTML(200, "new_blogpost", gin.H{
		"navigation": gin.H{
			"menu": c.Menu(),
			"headers": c.Headers().Elements,
		},
		"HttpMethod": "post",
		"Topics": helpers.Topics,
		"Created": time.Now().UTC().String(),

	})
}


func (c *Controller) MakeBlogPost(ctx *gin.Context) {
	var doc helpers.Document

	err := ctx.ShouldBind(&doc); if err != nil {
		ctx.HTML(500, "upload_status", gin.H{"UpdateMessage": "Update Failed!", "Color": "red"})
		return
	}
	rds := helpers.NewRedisClient(helpers.RedisConf{Addr: os.Getenv("REDIS_ADDR"), Port: os.Getenv("REDIS_PORT")})
	err = rds.AddDoc(doc); if err != nil {
		ctx.HTML(400, "upload_status", gin.H{"UpdateMessage": "Update Failed!", "Color": "red"})
		return
	}
	ctx.HTML(200, "upload_status", gin.H{"UpdateMessage": "Update Successful!", "Color": "green"})

}


func (c *Controller) ServeFileUpload(ctx *gin.Context) {
	ctx.HTML(200, "upload", gin.H{
		"navigation": gin.H{
			"menu": c.Menu(),
			"headers": c.Headers().Elements,
		},
	})
}



func (c *Controller) SaveFile(ctx *gin.Context) {
	file, _ := ctx.FormFile("file")

	// Upload the file to specific dst.
	err := ctx.SaveUploadedFile(file, fmt.Sprintf("%s/%s", helpers.GetImageStore(), file.Filename))
	if err != nil {
		ctx.HTML(400, "upload_status", gin.H{"UpdateMessage": "Update Failed!", "Color": "red"})
		return
	}

	ctx.HTML(200, "upload_status", gin.H{"UpdateMessage": "Update Successful!", "Color": "green"})
}
