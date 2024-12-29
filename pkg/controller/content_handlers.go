package controller

import (
	"time"
	"io"
	"bytes"
	"log"

	"git.aetherial.dev/aeth/keiji/pkg/helpers"
	"github.com/gin-gonic/gin"
)


func (c *Controller) ServeBlogDirectory(ctx *gin.Context) {
	ctx.HTML(200, "admin", gin.H{
		"navigation": gin.H{
			"menu": c.database.GetDropdownElements(),
			"headers": c.database.GetNavBarLinks(),
		},
		"Tables": c.database.GetAdminTables().Tables,

	})

}


func (c *Controller) GetBlogPostEditor(ctx *gin.Context) {
	post, exist := ctx.Params.Get("id")
	if !exist {
		ctx.JSON(404, map[string]string{
			"Error": "the requested file could not be found",
		})
		return
	}
	doc, err := c.database.GetDocument(helpers.Identifier(post))
	if err != nil {
		ctx.JSON(500, map[string]string{
			"Error": err.Error(),
		})
		return
	}
	ctx.HTML(200, "blogpost_editor", gin.H{
		"navigation": gin.H{
			"menu": c.database.GetDropdownElements(),
			"headers": c.database.GetNavBarLinks(),
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
	err = c.database.UpdateDocument(doc); if err != nil {
		ctx.HTML(400, "upload_status", gin.H{"UpdateMessage": "Update Failed!", "Color": "red"})
		return
	}
	ctx.HTML(200, "upload_status", gin.H{"UpdateMessage": "Update Successful!", "Color": "green"})

}



func (c *Controller) ServeNewBlogPage(ctx *gin.Context) {

	ctx.HTML(200, "new_blogpost", gin.H{
		"navigation": gin.H{
			"menu": c.database.GetDropdownElements(),
			"headers": c.database.GetNavBarLinks(),
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
	err = c.database.AddDocument(doc); if err != nil {
		ctx.HTML(400, "upload_status", gin.H{"UpdateMessage": "Update Failed!", "Color": "red"})
		return
	}
	ctx.HTML(200, "upload_status", gin.H{"UpdateMessage": "Update Successful!", "Color": "green"})

}


func (c *Controller) ServeFileUpload(ctx *gin.Context) {
	ctx.HTML(200, "upload", gin.H{
		"navigation": gin.H{
			"menu": c.database.GetDropdownElements(),
			"headers": c.database.GetNavBarLinks(),
		},
	})
}



func (c *Controller) SaveFile(ctx *gin.Context) {
	var img helpers.Image
	err := ctx.ShouldBind(&img); if err != nil {
		ctx.HTML(500, "upload_status", gin.H{"UpdateMessage": err, "Color": "red"})
		return
	}
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.HTML(500, "upload_status", gin.H{"UpdateMessage": err, "Color": "red"})
		return
	}
	fh, err := file.Open()
	if err != nil {
		ctx.HTML(500, "upload_status", gin.H{"UpdateMessage": err, "Color": "red"})
		return
	}
	fb := make([]byte, file.Size)
	var output bytes.Buffer
	for {
		n, err := fh.Read(fb)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		output.Write(fb[:n])
	}
	err = c.database.AddImage(fb, img.Title, img.Desc); if err != nil {
		ctx.HTML(500, "upload_status", gin.H{"UpdateMessage": err, "Color": "red"})
		return
	}

	ctx.HTML(200, "upload_status", gin.H{"UpdateMessage": "Update Successful!", "Color": "green"})
}
