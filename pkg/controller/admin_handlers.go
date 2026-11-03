package controller

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"git.aetherial.dev/aeth/keiji/pkg/auth"
	"git.aetherial.dev/aeth/keiji/pkg/storage"
	"github.com/gin-gonic/gin"
)

const AUTH_COOKIE_NAME = "X-Server-Auth"

// adminTables and adminImages preserve compatibility with legacy repositories.
func (c *Controller) adminTables() (storage.AdminPage, error) {
	if reader, ok := c.database.(storage.DocumentListReader); ok {
		return reader.GetAdminTablesWithError()
	}
	return c.database.GetAdminTables(), nil
}

func (c *Controller) adminImages() ([]storage.Image, error) {
	if reader, ok := c.database.(storage.DocumentListReader); ok {
		return reader.GetAllImagesWithError()
	}
	return c.database.GetAllImages(), nil
}

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
	page, err := c.adminTables()
	if err != nil {
		htmlListError(ctx, err)
		return
	}
	headers, menu, err := c.htmlNavigation()
	if err != nil {
		htmlListError(ctx, err)
		return
	}
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(AUTH_COOKIE_NAME, cookie, 1200, "/", "", !insecureCookies(), true)

	tables := page.Tables
	if tables == nil {
		tables = map[string][]storage.TableData{}
	}
	tables["modify"] = append(tables["modify"], storage.TableData{DisplayName: "Digital Art", Link: "/admin/digital"})
	ctx.HTML(http.StatusOK, "admin", gin.H{
		"navigation": gin.H{
			"headers": headers,
			"menu":    menu,
		},
		"Tables": tables,
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
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.HTML(400, "upload_status", gin.H{"UpdateMessage": err, "Color": "red"})
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
			ctx.HTML(500, "upload_status", gin.H{"UpdateMessage": err, "Color": "red"})
			return
		}
		output.Write(fb[:n])
	}

	var item storage.NavBarItem
	item = storage.NavBarItem{
		Link: file.Filename,
		Png:  fb,
	}
	err = ctx.ShouldBind(&item)
	if err != nil {
		ctx.HTML(400, "upload_status", gin.H{"UpdateMessage": err, "Color": "red"})
		return
	}
	err = c.database.AddNavbarItem(item)
	if err != nil {
		ctx.HTML(500, "upload_status", gin.H{"UpdateMessage": err, "Color": "red"})
		return
	}

	err = c.database.AddAsset(item.File.Filename, fb)
	if err != nil {
		ctx.HTML(500, "upload_status", gin.H{"UpdateMessage": err, "Color": "red"})
		return
	}

	ctx.HTML(200, "upload_status", gin.H{"UpdateMessage": "Addition Successful!", "Color": "green"})
}

/*
@Name RemoveNavbarItem
@Summary remove an entry from the navbar
@Tags admin
@Router /admin/navbar
*/
func (c *Controller) RemoveNavbarItem(ctx *gin.Context) {
	itemName, ok := ctx.Params.Get("item_name")
	if !ok {
		ctx.HTML(500, "upload_status", gin.H{"UpdateMessage": errors.New("No navbar name passed."), "Color": "red"})
		return
	}
	err := c.database.DeleteNavbarItem(storage.Identifier(itemName))
	if err != nil {
		ctx.HTML(500, "upload_status", gin.H{"UpdateMessage": err, "Color": "red"})
		return
	}
	ctx.HTML(200, "upload_status", gin.H{"UpdateMessage": "Deleted '" + itemName + "' Successfully!", "Color": "green"})

}

func (c *Controller) ServeNavbarModify(ctx *gin.Context) {
	navbar, menu, err := c.htmlNavigation()
	if err != nil {
		htmlListError(ctx, err)
		return
	}
	tableData := storage.AdminPage{Tables: map[string][]storage.TableData{}}
	for i := range navbar {
		tableData.Tables["items"] = append(tableData.Tables["items"],
			storage.TableData{
				DisplayName: navbar[i].Link,
				Link:        fmt.Sprintf("/admin/navbar/options/%s", navbar[i].Link),
			},
		)
	}
	ctx.HTML(200, "admin", gin.H{
		"navigation": gin.H{
			"menu":    menu,
			"headers": navbar,
		},
		"Tables": tableData.Tables,
	})

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

	page, err := c.adminTables()
	if err != nil {
		htmlListError(ctx, err)
		return
	}
	headers, menu, err := c.htmlNavigation()
	if err != nil {
		htmlListError(ctx, err)
		return
	}
	tables := page.Tables
	if tables == nil {
		tables = map[string][]storage.TableData{}
	}
	tables["modify"] = append(tables["modify"], storage.TableData{DisplayName: "Digital Art", Link: "/admin/digital"})
	ctx.HTML(http.StatusOK, "admin", gin.H{
		"navigation": gin.H{
			"headers": headers,
			"menu":    menu,
		},
		"Tables": tables,
	})

}

/*
Serves the admin panel with all of the documents in each blog category for editing
*/
func (c *Controller) ServeBlogDirectory(ctx *gin.Context) {
	tableData := storage.AdminPage{Tables: map[string][]storage.TableData{}}
	for i := range storage.Topics {
		docs, err := c.htmlDocuments(storage.Topics[i])
		if err != nil {
			htmlListError(ctx, err)
			return
		}
		for z := range docs {
			tableData.Tables[storage.Topics[i]] = append(tableData.Tables[storage.Topics[i]],
				storage.TableData{
					DisplayName: docs[z].Title,
					Link:        fmt.Sprintf("/admin/options/%s", docs[z].Ident),
				},
			)
		}
	}

	headers, menu, err := c.htmlNavigation()
	if err != nil {
		htmlListError(ctx, err)
		return
	}
	ctx.HTML(200, "admin", gin.H{
		"navigation": gin.H{
			"menu":    menu,
			"headers": headers,
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
	headers, menu, err := c.htmlNavigation()
	if err != nil {
		htmlListError(ctx, err)
		return
	}
	images, err := c.adminImages()
	if err != nil {
		htmlListError(ctx, err)
		return
	}
	ctx.HTML(200, "blogpost_editor", gin.H{
		"navigation": gin.H{
			"menu":    menu,
			"headers": headers,
		},
		"Ident":        doc.Ident,
		"Topics":       storage.Topics,
		"Title":        doc.Title,
		"DefaultTopic": doc.Category,
		"Created":      doc.Created,
		"Body":         doc.Body,
		"Images":       images,
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

	headers, menu, err := c.htmlNavigation()
	if err != nil {
		htmlListError(ctx, err)
		return
	}
	images, err := c.adminImages()
	if err != nil {
		htmlListError(ctx, err)
		return
	}
	ctx.HTML(200, "blogpost_editor", gin.H{
		"navigation": gin.H{
			"menu":    menu,
			"headers": headers,
		},
		"Post":    true,
		"Images":  images,
		"Topics":  storage.Topics,
		"Created": time.Now().UTC().String(),
	})
}

/*
Serving the new blogpost page. Serves the editor with the method to POST a new document
*/
func (c *Controller) ServeNewNavbarItem(ctx *gin.Context) {

	headers, menu, err := c.htmlNavigation()
	if err != nil {
		htmlListError(ctx, err)
		return
	}
	ctx.HTML(200, "navbar_editor", gin.H{
		"navigation": gin.H{
			"menu":    menu,
			"headers": headers,
		},
		"Post": true,
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
	headers, menu, err := c.htmlNavigation()
	if err != nil {
		htmlListError(ctx, err)
		return
	}
	ctx.HTML(200, "upload", gin.H{
		"navigation": gin.H{
			"menu":    menu,
			"headers": headers,
		},
		"Topics":       storage.ImageCategories,
		"DefaultTopic": storage.DIGITAL_ART,
	})
}

/*
Reciever for the page served to created a new visual media post
*/
// maxUploadSize limits the file itself; the request has additional room for form fields.
const maxUploadSize = 32 << 20
const maxUploadRequestSize = maxUploadSize + (1 << 20)

func (c *Controller) SaveFile(ctx *gin.Context) {
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxUploadRequestSize)
	defer func() {
		if ctx.Request.MultipartForm != nil {
			ctx.Request.MultipartForm.RemoveAll()
		}
	}()
	fail := func(status int, message string) {
		ctx.HTML(status, "upload_status", gin.H{"UpdateMessage": message, "Color": "red"})
	}
	var img storage.Image
	if err := ctx.ShouldBind(&img); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			fail(http.StatusRequestEntityTooLarge, "Upload too large")
		} else {
			fail(http.StatusBadRequest, "Invalid upload")
		}
		return
	}
	file, err := ctx.FormFile("file")
	if err != nil {
		fail(http.StatusBadRequest, "Missing upload file")
		return
	}
	if file.Size > maxUploadSize {
		fail(http.StatusRequestEntityTooLarge, "Upload too large")
		return
	}
	fh, err := file.Open()
	if err != nil {
		fail(http.StatusInternalServerError, "Could not open upload")
		return
	}
	defer fh.Close()
	data, err := io.ReadAll(io.LimitReader(fh, maxUploadSize+1))
	if err != nil {
		fail(http.StatusInternalServerError, "Could not read upload")
		return
	}
	if len(data) > maxUploadSize {
		fail(http.StatusRequestEntityTooLarge, "Upload too large")
		return
	}
	mediaType := detectMediaType(data)
	if mediaType == "" {
		fail(http.StatusUnsupportedMediaType, "Unsupported media type")
		return
	}
	if _, err = c.database.AddImage(data, img.Title, img.Desc, img.Category, mediaType); err != nil {
		fail(http.StatusInternalServerError, "Could not save upload")
		return
	}
	ctx.HTML(http.StatusOK, "upload_status", gin.H{"UpdateMessage": "Update Successful!", "Color": "green"})
}

// Serve the document deletion template
func (c *Controller) NavbarOptions(ctx *gin.Context) {
	id, found := ctx.Params.Get("id")
	if !found {
		ctx.HTML(400, "upload_status", gin.H{"UpdateMessage": "No ID selected!", "Color": "red"})
		return
	}

	ctx.HTML(200, "post_options", gin.H{
		"Link": fmt.Sprintf("/admin/navbar/%s", id),
	})

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

// Digital art administration is restricted by the /admin authentication middleware.
func (c *Controller) ListDigitalArt(ctx *gin.Context) {
	images, err := c.htmlImages(storage.DIGITAL_ART)
	if err != nil {
		htmlListError(ctx, err)
		return
	}
	ctx.HTML(http.StatusOK, "digital_art_admin_list", gin.H{"Images": images})
}

func (c *Controller) EditDigitalArt(ctx *gin.Context) {
	image, err := c.database.GetImage(storage.Identifier(ctx.Param("id")))
	if err != nil || image.Category != storage.DIGITAL_ART {
		ctx.String(http.StatusNotFound, "Digital art not found")
		return
	}
	ctx.HTML(http.StatusOK, "digital_art_admin_editor", image)
}

func (c *Controller) UpdateDigitalArt(ctx *gin.Context) {
	title := strings.TrimSpace(ctx.PostForm("title"))
	desc := ctx.PostForm("description")
	if title == "" {
		ctx.HTML(http.StatusBadRequest, "upload_status", gin.H{"UpdateMessage": "Title is required", "Color": "red"})
		return
	}
	err := c.database.UpdateDigitalArt(storage.Identifier(ctx.Param("id")), title, desc)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Could not update digital art"
		if errors.Is(err, storage.ErrNotExists) {
			status = http.StatusNotFound
			message = "Digital art not found"
		} else {
			log.Printf("update digital art failed: %v", err)
		}
		ctx.HTML(status, "upload_status", gin.H{"UpdateMessage": message, "Color": "red"})
		return
	}
	ctx.HTML(http.StatusOK, "upload_status", gin.H{"UpdateMessage": "Digital art updated", "Color": "green"})
}

// RemoveDigitalArt removes the gallery listing while preserving referenced media.
func (c *Controller) RemoveDigitalArt(ctx *gin.Context) {
	err := c.database.ArchiveDigitalArt(storage.Identifier(ctx.Param("id")))
	if err != nil {
		status := http.StatusInternalServerError
		message := "Could not remove digital art listing"
		if errors.Is(err, storage.ErrNotExists) {
			status = http.StatusNotFound
			message = "Digital art not found"
		} else {
			log.Printf("remove digital art listing failed: %v", err)
		}
		ctx.HTML(status, "upload_status", gin.H{"UpdateMessage": message, "Color": "red"})
		return
	}
	ctx.HTML(http.StatusOK, "upload_status", gin.H{"UpdateMessage": "Listing removed. The media file remains available for posts.", "Color": "green"})
}
