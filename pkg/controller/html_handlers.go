package controller

import (
	"html/template"
	"log"
	"math/rand"
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

// These helpers retain compatibility with older DocumentIO implementations while
// allowing repositories that implement DocumentListReader to report failures.
func (c *Controller) htmlDocuments(category string) ([]storage.Document, error) {
	if reader, ok := c.database.(storage.DocumentListReader); ok {
		return reader.GetByCategoryWithError(category)
	}
	return c.database.GetByCategory(category), nil
}

func (c *Controller) htmlImages(category string) ([]storage.Image, error) {
	if reader, ok := c.database.(storage.DocumentListReader); ok {
		return reader.GetImagesByCategoryWithError(category)
	}
	return c.database.GetImagesByCategory(category), nil
}

func (c *Controller) htmlNavigation() ([]storage.NavBarItem, []storage.LinkPair, error) {
	if reader, ok := c.database.(storage.DocumentListReader); ok {
		headers, err := reader.GetNavBarLinksWithError()
		if err != nil {
			return nil, nil, err
		}
		menu, err := reader.GetDropdownElementsWithError()
		return headers, menu, err
	}
	return c.database.GetNavBarLinks(), c.database.GetDropdownElements(), nil
}

func htmlListError(ctx *gin.Context, err error) {
	log.Printf("list query failed: %v", err)
	ctx.JSON(http.StatusInternalServerError, gin.H{"Error": "Could not load content"})
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
	headers, menu, err := c.htmlNavigation()
	if err != nil {
		htmlListError(ctx, err)
		return
	}
	ctx.HTML(http.StatusOK, "blogpost", gin.H{
		"navigation": gin.H{
			"headers": headers,
		},
		"Title":   doc.Title,
		"Ident":   doc.Ident,
		"Created": doc.Created,
		"Body":    template.HTML(MdToHTML([]byte(doc.Body))),
		"menu":    menu,
	})

}

// @Name ServeBlogHome
// @Summary serves the HTML file for the blog post homepage
// @Tags webpages
// @Router / [get]
func (c *Controller) ServeHome(ctx *gin.Context) {
	home, err := c.htmlDocuments(storage.HOMEPAGE)
	if err != nil {
		htmlListError(ctx, err)
		return
	}
	headers, menu, err := c.htmlNavigation()
	if err != nil {
		htmlListError(ctx, err)
		return
	}
	var content storage.Document
	if len(home) == 0 {
		content = storage.Document{
			Body: "Under construction. Sry :(",
		}
	} else {
		index := rand.Intn(len(home))
		content = home[index]
	}
	ctx.HTML(http.StatusOK, "home", gin.H{
		"navigation": gin.H{
			"headers": headers,
		},
		"menu":    menu,
		"default": content,
	})
}

// @Name ServeBlog
// @Summary serves the HTML for written post listings
// @Tags webpages
// @Router /blog [get]
func (c *Controller) ServeBlog(ctx *gin.Context) {
	docs, err := c.htmlDocuments(storage.BLOG)
	if err != nil {
		htmlListError(ctx, err)
		return
	}
	ctx.HTML(http.StatusOK, "writing", docs)
}

// @Name ServeCreative
// @Summary serves the HTML for the creative writing listings
// @Tags webpages
// @Router /creative [get]
func (c *Controller) ServeCreative(ctx *gin.Context) {
	docs, err := c.htmlDocuments(storage.CREATIVE)
	if err != nil {
		htmlListError(ctx, err)
		return
	}
	ctx.HTML(http.StatusOK, "writing", docs)
}

// @Name ServeDigitalArt
// @Summary serves the HTML file for the digital art homepage
// @Tags webpages
// @Router /digital [get]
func (c *Controller) ServeDigitalArt(ctx *gin.Context) {
	images, err := c.htmlImages(storage.DIGITAL_ART)
	if err != nil {
		htmlListError(ctx, err)
		return
	}
	headers, menu, err := c.htmlNavigation()
	if err != nil {
		htmlListError(ctx, err)
		return
	}
	ctx.HTML(http.StatusOK, "digital_art", gin.H{
		"navigation": gin.H{
			"headers": headers,
		},
		"images": images,
		"menu":   menu,
	})
}
