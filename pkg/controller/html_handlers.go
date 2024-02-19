package controller

import (
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
			"menu": c.Menu(),
			"headers": c.Headers().Elements,
		},
		"title": doc.Ident,
		"Ident": doc.Ident,
		"Created": doc.Created,
		"Body": doc.Body,

	})

}

// @Name ServeBlogHome
// @Summary serves the HTML file for the blog post homepage
// @Tags webpages
// @Router /blog [get]
func (c *Controller) ServeBlogHome(ctx *gin.Context) {
	docs, err := helpers.GetAllDocuments(helpers.BLOG, c.RedisConfig)
	if err != nil {
		ctx.JSON(500, map[string]string{
			"Error getting docs": err.Error(),
		})
		return
	}
	ctx.HTML(http.StatusOK, "home", gin.H{
		"navigation": gin.H{
			"menu": c.Menu(),
			"headers": c.Headers().Elements,
		},
		"listings": docs,
	})
}


// @Name ServeHtml
// @Summary serves HTML files out of the HTML directory
// @Tags webpages
// @Router /home [get]
func (c *Controller) ServeHome(ctx *gin.Context) {
	docs, err := helpers.GetAllDocuments(helpers.BLOG, c.RedisConfig)
	if err != nil {
		ctx.JSON(500, map[string]string{
			"Error getting docs": err.Error(),
		})
		return
	}
	ctx.HTML(http.StatusOK, "home", gin.H{
		"navigation": gin.H{
			"menu": c.Menu(),
			"headers": c.Headers().Elements,
		},
		"listings": docs,
	})
}

// @Name ServeCreativeWriting
// @Summary serves the HTML file for the creative writing homepage
// @Tags webpages
// @Router /creative [get]
func (c *Controller) ServeCreativeWriting(ctx *gin.Context) {
	docs, err := helpers.GetAllDocuments(helpers.CREATIVE, c.RedisConfig)
	if err != nil {
		ctx.JSON(500, map[string]string{
			"Error getting docs": err.Error(),
		})
		return
	}
	ctx.HTML(http.StatusOK, "home", gin.H{
		"navigation": gin.H{
			"menu": c.Menu(),
			"headers": c.Headers().Elements,
		},
		"listings": docs,
	})

}

// @Name ServeTechnicalWriteups
// @Summary serves the HTML file for the technical writeup homepage
// @Tags webpages
// @Router /writeups [get]
func (c *Controller) ServeTechnicalWriteups(ctx *gin.Context) {
	docs, err := helpers.GetAllDocuments(helpers.TECHNICAL, c.RedisConfig)
	if err != nil {
		ctx.JSON(500, map[string]string{
			"Error getting docs": err.Error(),
		})
		return
	}
	ctx.HTML(http.StatusOK, "home", gin.H{
		"navigation": gin.H{
			"menu": c.Menu(),
			"headers": c.Headers().Elements,
		},
		"listings": docs,
	})

}

// @Name ServeDigitalArt
// @Summary serves the HTML file for the digital art homepage
// @Tags webpages
// @Router /digital [get]
func (c *Controller) ServeDigitalArt(ctx *gin.Context) {
	fnames, err := helpers.GetImagePaths(4, 0)
	if err != nil {
		ctx.HTML(http.StatusInternalServerError, "unhandled_error",
		gin.H{
			"StatusCode": http.StatusInternalServerError,
			"Reason": err.Error(),
		},
	)
	return
	}
	ctx.HTML(http.StatusOK, "digital_art", gin.H{
		"navigation": gin.H{
			"menu": c.Menu(),
			"headers": c.Headers().Elements,
		},
		"images": fnames,
	})
}
