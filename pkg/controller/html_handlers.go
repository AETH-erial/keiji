package controller

import (
	"html/template"
	"net/http"

	"git.aetherial.dev/aeth/keiji/pkg/storage"
	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

/*
convert markdown to html

	:param md: the byte array containing the Markdown to convert
*/
func MdToHTML(md []byte) []byte {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

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
	doc, err := c.database.GetDocument(storage.Identifier(post))
	if err != nil {
		ctx.JSON(500, map[string]string{
			"Error": err.Error(),
		})
		return
	}
	if doc.Category == storage.CONFIGURATION {
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
		"Body":    template.HTML(MdToHTML([]byte(doc.Body))),
		"menu":    c.database.GetDropdownElements(),
	})

}

// @Name ServeBlogHome
// @Summary serves the HTML file for the blog post homepage
// @Tags webpages
// @Router / [get]
func (c *Controller) ServeHome(ctx *gin.Context) {
	home := c.database.GetByCategory(storage.HOMEPAGE)
	var content storage.Document
	if len(home) == 0 {
		content = storage.Document{
			Body: "Under construction. Sry :(",
		}
	} else {
		content = home[0]
	}
	ctx.HTML(http.StatusOK, "home", gin.H{
		"navigation": gin.H{
			"headers": c.database.GetNavBarLinks(),
		},
		"menu":    c.database.GetDropdownElements(),
		"default": content,
	})
}

// @Name ServeBlog
// @Summary serves the HTML for written post listings
// @Tags webpages
// @Router /blog [get]
func (c *Controller) ServeBlog(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "writing", c.database.GetByCategory(storage.BLOG))
}

// @Name ServeCreative
// @Summary serves the HTML for the creative writing listings
// @Tags webpages
// @Router /creative [get]
func (c *Controller) ServeCreative(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "writing", c.database.GetByCategory(storage.CREATIVE))
}

// @Name ServeDigitalArt
// @Summary serves the HTML file for the digital art homepage
// @Tags webpages
// @Router /digital [get]
func (c *Controller) ServeDigitalArt(ctx *gin.Context) {
	images := c.database.GetAllImages()
	ctx.HTML(http.StatusOK, "digital_art", gin.H{
		"navigation": gin.H{
			"headers": c.database.GetNavBarLinks(),
		},
		"images": images,
		"menu":   c.database.GetDropdownElements(),
	})
}
