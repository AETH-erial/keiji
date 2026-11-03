package controller

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"git.aetherial.dev/aeth/keiji/pkg/storage"
	"github.com/gin-gonic/gin"
)

// @Name ServeImage
// @Summary serves image from the image store
// @Tags cdn
// @Router /images/{file} [get]
func (c *Controller) ServeImage(ctx *gin.Context) {
	id, exists := ctx.Params.Get("file")
	if !exists || id == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"Error": "the requested file could not be found"})
		return
	}
	var image storage.Image
	var content io.ReadSeeker
	var file storage.ReadSeekCloser
	var err error
	if opener, ok := c.database.(interface {
		OpenImage(storage.Identifier) (storage.Image, storage.ReadSeekCloser, error)
	}); ok {
		image, file, err = opener.OpenImage(storage.Identifier(id))
		if err == nil && file == nil {
			err = errors.New("image stream is nil")
		}
		if file != nil {
			defer file.Close()
			content = file
		}
	} else {
		// Compatibility for repositories that only implement DocumentIO.
		image, err = c.database.GetImage(storage.Identifier(id))
		content = bytes.NewReader(image.Data)
	}
	if err != nil {
		if errors.Is(err, storage.ErrNotExists) || errors.Is(err, os.ErrNotExist) {
			ctx.JSON(http.StatusNotFound, gin.H{"Error": "the requested file could not be found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"Error": "Could not serve the requested file"})
		}
		return
	}
	mediaType := image.MediaType
	if mediaType == "" {
		mediaType = "application/octet-stream"
	}
	ctx.Header("Content-Type", mediaType)
	ctx.Header("X-Content-Type-Options", "nosniff")
	http.ServeContent(ctx.Writer, ctx.Request, "", time.Time{}, content)
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
	var assets []storage.Asset
	if reader, ok := c.database.(storage.DocumentListReader); ok {
		var err error
		assets, err = reader.GetAssetsWithError()
		if err != nil {
			htmlListError(ctx, err)
			return
		}
	} else {
		assets = c.database.GetAssets()
	}
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
	case fext == "json" || fext == "map":
		ctype = "application/json"
	case fext == "png":
		ctype = "image/png"
	default:
		ctype = "text"
	}
	fh, err := c.FileIO.Open(path.Join("cdn", f))
	if err != nil {
		ctx.JSON(500, map[string]string{
			"Error": "Could not serve the requested file",
			"msg":   err.Error(),
		})
		return
	}
	defer fh.Close()
	b, err := io.ReadAll(fh)
	if err != nil {
		ctx.JSON(500, map[string]string{
			"Error": "Could not serve the requested file",
			"msg":   err.Error(),
		})
		return

	}
	ctx.Data(200, ctype, b)
}
