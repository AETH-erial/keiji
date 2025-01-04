package controller

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"git.aetherial.dev/aeth/keiji/pkg/auth"
	"git.aetherial.dev/aeth/keiji/pkg/storage"
	"github.com/gin-gonic/gin"
)

const AUTH_COOKIE_NAME = "X-Server-Auth"

// @Name ServeLogin
// @Summary serves the HTML login page
// @Tags admin
// @Router /login [get]
func (c *Controller) ServeLogin(ctx *gin.Context) {
	cookie, _ := ctx.Cookie(AUTH_COOKIE_NAME)
	if c.Cache.Read(cookie) {
		ctx.Redirect(302, "/home")
	}
	ctx.HTML(http.StatusOK, "login", gin.H{
		"heading": "aetherial.dev login",
	})

}

// @Name Auth
// @Summary serves recieves admin user and pass, sets a cookie
// @Tags admin
// @Param cred body storage.Credentials true "Admin Credentials"
// @Router /login [post]
func (c *Controller) Auth(ctx *gin.Context) {

	var cred auth.Credentials

	err := ctx.ShouldBind(&cred)
	if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}
	cookie, err := auth.Authorize(&cred, c.Cache, c.AuthSource)
	if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}
	ctx.SetCookie(AUTH_COOKIE_NAME, cookie, 3600, "/", c.Domain, false, false)

	ctx.HTML(http.StatusOK, "admin", gin.H{
		"navigation": gin.H{
			"headers": c.database.GetNavBarLinks(),
			"menu":    c.database.GetDropdownElements(),
		},
		"Tables": c.database.GetAdminTables().Tables,
	})

}

/*
@Name AddAdminTableEntry
@Summary add an entry to the admin table
@Tags admin
@Router /admin/panel
*/
func (c *Controller) AddAdminTableEntry(ctx *gin.Context) {
	tables := make(map[string][]storage.TableData)
	adminPage := storage.AdminPage{Tables: tables}
	err := ctx.ShouldBind(&adminPage)
	if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}
	for category := range adminPage.Tables {
		for entry := range adminPage.Tables[category] {
			err := c.database.AddAdminTableEntry(adminPage.Tables[category][entry], category)
			if err != nil {
				ctx.JSON(400, map[string]string{
					"Error": err.Error(),
				})
				return
			}

		}
	}
	ctx.Data(200, "text", []byte("Categories populated."))
}

/*
@Name AddMenuItem
@Summary add an entry to the sidebar menu
@Tags admin
@Router /admin/menu
*/
func (c *Controller) AddMenuItem(ctx *gin.Context) {
	var item storage.LinkPair
	err := ctx.ShouldBind(&item)
	if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}
	err = c.database.AddMenuItem(item)
	if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}
	ctx.Data(200, "text", []byte("menu item added."))
}

/*
@Name AddNavbarItem
@Summary add an entry to the navbar
@Tags admin
@Router /admin/navbar
*/
func (c *Controller) AddNavbarItem(ctx *gin.Context) {

	var item storage.NavBarItem
	err := ctx.ShouldBind(&item)
	if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}
	err = c.database.AddNavbarItem(item)
	if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}

	err = c.database.AddAsset(item.Link, item.Png)
	if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}

	ctx.Data(200, "text", []byte("navbar item added."))
}

/*
@Name AddAsset
@Summary add an asset to the db
@Tags admin
@Router /admin/assets
*/
func (c *Controller) AddAsset(ctx *gin.Context) {
	var item storage.Asset
	err := ctx.ShouldBind(&item)
	if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}
	err = c.database.AddAsset(item.Name, item.Data)
	if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}

}

// @Name AdminPanel
// @Summary serve the admin panel page
// @Tags admin
// @Router /admin/panel [get]
func (c *Controller) AdminPanel(ctx *gin.Context) {

	ctx.HTML(http.StatusOK, "admin", gin.H{
		"navigation": gin.H{
			"headers": c.database.GetNavBarLinks(),
			"menu":    c.database.GetDropdownElements(),
		},
		"Tables": c.database.GetAdminTables().Tables,
	})

}

/*
Serves the admin panel with all of the documents in each blog category for editing
*/
func (c *Controller) ServeBlogDirectory(ctx *gin.Context) {
	tableData := storage.AdminPage{Tables: map[string][]storage.TableData{}}
	for i := range storage.Topics {
		docs := c.database.GetByCategory(storage.Topics[i])
		for z := range docs {
			tableData.Tables[storage.Topics[i]] = append(tableData.Tables[storage.Topics[i]],
				storage.TableData{
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
	doc, err := c.database.GetDocument(storage.Identifier(post))
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
		"Topics":       storage.Topics,
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
	var doc storage.Document

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
		"Topics":  storage.Topics,
		"Created": time.Now().UTC().String(),
	})
}

/*
Reciever for the ServeNewBlogPage UI screen. Adds a new document to the database
*/
func (c *Controller) MakeBlogPost(ctx *gin.Context) {
	var doc storage.Document
	err := ctx.ShouldBind(&doc)
	if err != nil {
		ctx.HTML(500, "upload_status", gin.H{"UpdateMessage": "Update Failed!", "Color": "red"})
		return
	}
	_, err = c.database.AddDocument(doc)
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
	var img storage.Image
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
	_, err = c.database.AddImage(fb, img.Title, img.Desc)
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
	err := c.database.DeleteDocument(storage.Identifier(id))
	if err != nil {
		ctx.HTML(500, "upload_status", gin.H{"UpdateMessage": "Delete Failed!", "Color": "red"})
		return
	}
	ctx.HTML(200, "upload_status", gin.H{"UpdateMessage": "Delete Successful!", "Color": "green"})

}
