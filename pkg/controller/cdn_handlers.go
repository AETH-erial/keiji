package controller

import (
	"fmt"
	"os"

	"git.aetherial.dev/aeth/keiji/pkg/helpers"
	"github.com/gin-gonic/gin"
)

// @Name ServeCss
// @Summary serves css files from the web root directory
// @Tags cdn
// @Param file path string true "The CSS file to serve to the client"
// @Router /api/v1/style/{file} [get]
func (c *Controller) ServeCss(ctx *gin.Context) {
	f, exist := ctx.Params.Get("file")
	if !exist {
		ctx.JSON(404, map[string]string{
			"Error": "the requested file could not be found",
		})
	}
	css := fmt.Sprintf("%s/css/bootstrap-5.0.2/dist/css/%s", c.WebRoot, f)
	b, err := os.ReadFile(css)
	if err != nil {
		ctx.JSON(500, map[string]string{
			"Error": "Could not serve the requested file",
			"msg":   err.Error(),
		})
	}
	ctx.Data(200, "text/css", b)

}

// @Name ServeJs
// @Summary serves js files from the web root directory
// @Tags cdn
// @Param file path string true "The Javascript file to serve to the client"
// @Router /api/v1/js/{file} [get]
func (c *Controller) ServeJs(ctx *gin.Context) {
	f, exist := ctx.Params.Get("file")
	if !exist {
		ctx.JSON(404, map[string]string{
			"Error": "the requested file could not be found",
		})
	}
	css := fmt.Sprintf("%s/css/bootstrap-5.0.2/dist/js/%s", c.WebRoot, f)
	b, err := os.ReadFile(css)
	if err != nil {
		ctx.JSON(500, map[string]string{
			"Error": "Could not serve the requested file",
			"msg":   err.Error(),
		})
	}
	ctx.Data(200, "text/javascript", b)

}

// @Name ServeMdbCss
// @Summary serves some mdb assets
// @Tags cdn
// @Param file path string true "The CSS file to serve to the client"
// @Router /api/v1/style/mdb/{file} [get]
func (c *Controller) ServeMdbCss(ctx *gin.Context) {
	f, exist := ctx.Params.Get("file")
	if !exist {
		ctx.JSON(404, map[string]string{
			"Error": "the requested file could not be found",
		})
	}
	css := fmt.Sprintf("%s/css/MDB5-STANDARD-UI-KIT-Free-7.1.0/css/%s", c.WebRoot, f)
	b, err := os.ReadFile(css)
	if err != nil {
		ctx.JSON(500, map[string]string{
			"Error": "Could not serve the requested file",
			"msg":   err.Error(),
		})
	}
	ctx.Data(200, "text/css", b)

}

// @Name ServeAsset
// @Summary serves assets to put in a webpage
// @Tags cdn
// @Router /assets/{file} [get]
func (c *Controller) ServeAsset(ctx *gin.Context) {
	f, exist := ctx.Params.Get("file")
	if !exist {
		ctx.JSON(404, map[string]string{
			"Error": "the requested file could not be found",
		})
	}
	css := fmt.Sprintf("%s/assets/%s", c.WebRoot, f)
	b, err := os.ReadFile(css)
	if err != nil {
		ctx.JSON(500, map[string]string{
			"Error": "Could not serve the requested file",
			"msg":   err.Error(),
		})
	}
	ctx.Data(200, "image/jpeg", b)
}

// @Name ServeImage
// @Summary serves image from the image store
// @Tags cdn
// @Router /images/{file} [get]
func (c *Controller) ServeImage(ctx *gin.Context) {
	f, exist := ctx.Params.Get("file")
	if !exist {
		ctx.JSON(404, map[string]string{
			"Error": "the requested file could not be found",
		})
	}
	css := fmt.Sprintf("%s/%s", helpers.GetImageStore(), f)
	b, err := os.ReadFile(css)
	if err != nil {
		ctx.JSON(500, map[string]string{
			"Error": "Could not serve the requested file",
			"msg":   err.Error(),
		})
	}
	ctx.Data(200, "image/jpeg", b)
}

