package controller

import (
	"html/template"
	"net/http"
	

	"git.aetherial.dev/aeth/keiji/pkg/helpers"
	"github.com/gin-gonic/gin"
)

// @Name ServePost
// @Summary serves HTML files out of the HTML directory
// @Tags webpages
// @Router /writing/:id [get]
func (c *Controller) ServePost(ctx *gin.Context) {
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
	if doc.Category == helpers.CONFIGURATION {
		ctx.Status(404)
		return
	}
	ctx.HTML(http.StatusOK, "blogpost", gin.H{
		"navigation": gin.H{
			"headers": c.database.GetNavBarLinks(),
		},
		"Title":   doc.Title,
		"Ident":   doc.Ident,
		"Created": doc.Created,
		"Body":    template.HTML(helpers.MdToHTML([]byte(doc.Body))),
		"menu":    c.database.GetDropdownElements(),
	})

}

// @Name ServeBlogHome
// @Summary serves the HTML file for the blog post homepage
// @Tags webpages
// @Router / [get]
func (c *Controller) ServeHome(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "home", gin.H{
		"navigation": gin.H{
			"headers": c.database.GetNavBarLinks(),
		},
		"menu":     c.database.GetDropdownElements(),
		"listings": c.database.GetByCategory(helpers.BLOG),
	})
}

// @Name ServeBlog
// @Summary serves the HTML for written post listings
// @Tags webpages
// @Router /blog [get]
func (c *Controller) ServeBlog(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "writing", c.database.GetByCategory(helpers.BLOG))
}


// @Name ServeDigitalArt
// @Summary serves the HTML file for the digital art homepage
// @Tags webpages
// @Router /digital [get]
func (c *Controller) ServeDigitalArt(ctx *gin.Context) {
	images := c.database.GetAllImages()
	/*
	if err != nil {
		ctx.HTML(http.StatusInternalServerError, "unhandled_error",
			gin.H{
				"StatusCode": http.StatusInternalServerError,
				"Reason":     err.Error(),
			},
		)
		return
	}
	*/
	ctx.HTML(http.StatusOK, "digital_art", gin.H{
		"navigation": gin.H{
			"headers": c.database.GetNavBarLinks(),
		},
		"images": images,
		"menu":   c.database.GetDropdownElements(),
	})
}
