package routes

import (
	"io/fs"

	"git.aetherial.dev/aeth/keiji/pkg/controller"
	"git.aetherial.dev/aeth/keiji/pkg/helpers"
	"github.com/gin-gonic/gin"
)

func Register(e *gin.Engine, domain string, redisPort string, redisAddr string, database helpers.DocumentIO, files fs.FS) {
	c := controller.NewController(domain, redisPort, redisAddr, database, files)
	web := e.Group("")
	web.GET("/", c.ServeHome)
	web.GET("/blog", c.ServeBlog)
	web.GET("/digital", c.ServeDigitalArt)
	web.GET("/writing/:id", c.ServePost)
	web.GET("/login", c.ServeLogin)
	web.POST("/login", c.Auth)

	cdn := e.Group("/api/v1")
	cdn.GET("/images/:file", c.ServeImage)
	cdn.GET("/cdn/:file", c.ServeGeneric)
	cdn.GET("assets/:file", c.ServeAsset)

	priv := e.Group("/admin")
	priv.Use(c.IsAuthenticated)
	priv.GET("/upload", c.ServeFileUpload)
	priv.POST("/upload", c.SaveFile)
	priv.GET("/panel", c.AdminPanel)
	priv.POST("/panel", c.AddAdminTableEntry)
	priv.POST("/menu", c.AddMenuItem)
	priv.POST("/navbar", c.AddNavbarItem)
	priv.POST("/images/upload", c.SaveFile)
	priv.GET("/posts/:id", c.GetBlogPostEditor)
	priv.POST("/posts", c.MakeBlogPost)
	priv.GET("/posts/all", c.ServeBlogDirectory)
	priv.GET("/posts", c.ServeNewBlogPage)
	priv.PATCH("/posts", c.UpdateBlogPost)

}
