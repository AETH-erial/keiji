package controller

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

	"git.aetherial.dev/aeth/keiji/pkg/helpers"
	"github.com/gin-gonic/gin"
)

/*
Serves the admin panel with all of the documents in each blog category for editing
*/
func (c *Controller) ServeBlogDirectory(ctx *gin.Context) {
	tableData := helpers.AdminPage{Tables: map[string][]helpers.TableData{}}
	for i := range helpers.Topics {
		docs := c.database.GetByCategory(helpers.Topics[i])
		for z := range docs {
			tableData.Tables[helpers.Topics[i]] = append(tableData.Tables[helpers.Topics[i]],
				helpers.TableData{
					DisplayName: docs[z].Title,
					Link:        fmt.Sprintf("/admin/options/%s", docs[z].Ident),
				},
			)
		}
	}

	ctx.HTML(200, "admin", gin.H{
		"navigation": gin.H{
			"menu":    c.database.GetDropdownElements(),
			"headers": c.database.GetNavBarLinks(),
		},
		"Tables": tableData.Tables,
	})

}

/*
Serves the blogpost editor with the submit button set to PATCH a document
*/
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
			"menu":    c.database.GetDropdownElements(),
			"headers": c.database.GetNavBarLinks(),
		},
		"Ident":        doc.Ident,
		"Topics":       helpers.Topics,
		"Title":        doc.Title,
		"DefaultTopic": doc.Category,
		"Created":      doc.Created,
		"Body":         doc.Body,
	})
}

/*
Update an existing blog post
*/
func (c *Controller) UpdateBlogPost(ctx *gin.Context) {
	var doc helpers.Document

	err := ctx.ShouldBind(&doc)
	if err != nil {
		ctx.HTML(500, "upload_status", gin.H{"UpdateMessage": "Update Failed!", "Color": "red"})
		return
	}
	err = c.database.UpdateDocument(doc)
	if err != nil {
		ctx.HTML(400, "upload_status", gin.H{"UpdateMessage": "Update Failed!", "Color": "red"})
		return
	}
	ctx.HTML(200, "upload_status", gin.H{"UpdateMessage": "Update Successful!", "Color": "green"})

}

/*
Serving the new blogpost page. Serves the editor with the method to POST a new document
*/
func (c *Controller) ServeNewBlogPage(ctx *gin.Context) {

	ctx.HTML(200, "blogpost_editor", gin.H{
		"navigation": gin.H{
			"menu":    c.database.GetDropdownElements(),
			"headers": c.database.GetNavBarLinks(),
		},
		"Post":    true,
		"Topics":  helpers.Topics,
		"Created": time.Now().UTC().String(),
	})
}

/*
Reciever for the ServeNewBlogPage UI screen. Adds a new document to the database
*/
func (c *Controller) MakeBlogPost(ctx *gin.Context) {
	var doc helpers.Document
	err := ctx.ShouldBind(&doc)
	if err != nil {
		ctx.HTML(500, "upload_status", gin.H{"UpdateMessage": "Update Failed!", "Color": "red"})
		return
	}
	err = c.database.AddDocument(doc)
	if err != nil {
		ctx.HTML(400, "upload_status", gin.H{"UpdateMessage": "Update Failed!", "Color": "red"})
		return
	}
	ctx.HTML(200, "upload_status", gin.H{"UpdateMessage": "Update Successful!", "Color": "green"})

}

/*
Serves the HTML page for a new visual media post
*/
func (c *Controller) ServeFileUpload(ctx *gin.Context) {
	ctx.HTML(200, "upload", gin.H{
		"navigation": gin.H{
			"menu":    c.database.GetDropdownElements(),
			"headers": c.database.GetNavBarLinks(),
		},
	})
}

/*
Reciever for the page served to created a new visual media post
*/
func (c *Controller) SaveFile(ctx *gin.Context) {
	var img helpers.Image
	err := ctx.ShouldBind(&img)
	if err != nil {
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
	err = c.database.AddImage(fb, img.Title, img.Desc)
	if err != nil {
		ctx.HTML(500, "upload_status", gin.H{"UpdateMessage": err, "Color": "red"})
		return
	}

	ctx.HTML(200, "upload_status", gin.H{"UpdateMessage": "Update Successful!", "Color": "green"})
}

// Serve the document deletion template
func (c *Controller) PostOptions(ctx *gin.Context) {
	id, found := ctx.Params.Get("id")
	if !found {
		ctx.HTML(400, "upload_status", gin.H{"UpdateMessage": "No ID selected!", "Color": "red"})
		return
	}

	ctx.HTML(200, "post_options", gin.H{
		"Link": fmt.Sprintf("/admin/posts/%s", id),
	})

}

/*
Delete a document from the database
*/
func (c *Controller) DeleteDocument(ctx *gin.Context) {
	id, found := ctx.Params.Get("id")
	if !found {
		ctx.HTML(500, "upload_status", gin.H{"UpdateMessage": "No ID passed!", "Color": "red"})
		return

	}
	err := c.database.DeleteDocument(helpers.Identifier(id))
	if err != nil {
		ctx.HTML(500, "upload_status", gin.H{"UpdateMessage": "Delete Failed!", "Color": "red"})
		return
	}
	ctx.HTML(200, "upload_status", gin.H{"UpdateMessage": "Delete Successful!", "Color": "green"})

}
