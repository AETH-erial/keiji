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
// @Router /writing/:post-name [get]
func (c *Controller) ServePost(ctx *gin.Context) {
	rds := helpers.NewRedisClient(c.RedisConfig)
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
	if doc.Category == helpers.CONFIGURATION {
		ctx.Status(404)
		return
	}
	ctx.HTML(http.StatusOK, "blogpost", gin.H{
		"navigation": gin.H{
			"headers": c.database.GetNavBarLinks(),
		},
		"title":   doc.Ident,
		"Ident":   doc.Ident,
		"Created": doc.Created,
		"Body":    template.HTML(helpers.MdToHTML([]byte(doc.Body))),
		"menu":    c.database.GetDropdownElements(),
	})

}

// @Name ServeBlogHome
// @Summary serves the HTML file for the blog post homepage
// @Tags webpages
// @Router /blog [get]
func (c *Controller) ServeBlogHome(ctx *gin.Context) {
	docs := c.database.GetByCategory(helpers.BLOG)
	ctx.HTML(http.StatusOK, "home", gin.H{
		"navigation": gin.H{
			"headers": c.database.GetNavBarLinks(),
		},
		"listings": docs,
		"menu":     c.database.GetDropdownElements(),
	})
}

// @Name ServeDigitalArt
// @Summary serves the HTML file for the digital art homepage
// @Tags webpages
// @Router /digital [get]
func (c *Controller) ServeDigitalArt(ctx *gin.Context) {
	rds := helpers.NewRedisClient(c.RedisConfig)
	fnames, err := helpers.GetImageData(rds)
	if err != nil {
		ctx.HTML(http.StatusInternalServerError, "unhandled_error",
			gin.H{
				"StatusCode": http.StatusInternalServerError,
				"Reason":     err.Error(),
			},
		)
		return
	}
	ctx.HTML(http.StatusOK, "digital_art", gin.H{
		"navigation": gin.H{
			"headers": c.database.GetNavBarLinks(),
		},
		"images": fnames,
		"menu":   c.database.GetDropdownElements(),
	})
}
