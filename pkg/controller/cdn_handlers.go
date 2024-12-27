package controller

import (
	"fmt"
	"os"
	"path"
	"strings"
	"net/http"

	"git.aetherial.dev/aeth/keiji/pkg/helpers"
	"github.com/gin-gonic/gin"
)

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

// @Name ServeAsset
// @Summary serves file from the html file
// @Tags cdn
// @Router /api/v1/assets/{file} [get]
func (c *Controller) ServeAsset(ctx *gin.Context) {
	f, exist := ctx.Params.Get("file")
	if !exist {
		ctx.JSON(404, map[string]string{
			"Error": "the requested file could not be found",
		})
		return
	}
	assets := c.database.GetAssets()
	for i := range assets {
		if strings.Contains(assets[i].Name, f) {
			ctx.Data(200, "image/png", assets[i].Data)
			return 
		}
	}
	ctx.Data(http.StatusNotFound, "text", []byte("Couldnt find the image requested."))

}



// @Name ServeGeneric
// @Summary serves file from the html file
// @Tags cdn
// @Router /api/v1/cdn/{file} [get]
func (c *Controller) ServeGeneric(ctx *gin.Context) {
	f, exist := ctx.Params.Get("file")
	if !exist {
		ctx.JSON(404, map[string]string{
			"Error": "the requested file could not be found",
		})
		return
	}
	fext := strings.Split(f, ".")[len(strings.Split(f, "."))-1]
	var ctype string
	switch {
	case fext == "css":
		ctype = "text/css"
	case fext == "js":
		ctype = "text/javascript"
	case fext == "json":
		ctype = "application/json"
	case fext == "png":
		ctype = "image/png"
	default:
		ctype = "text"
	}
	b, err := os.ReadFile(path.Join(c.WebRoot, f))
	if err != nil {
		ctx.JSON(500, map[string]string{
			"Error": "Could not serve the requested file",
			"msg":   err.Error(),
		})
		return
	}
	ctx.Data(200, ctype, b)
}
