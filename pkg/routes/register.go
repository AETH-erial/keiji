package routes

import (
	"git.aetherial.dev/aeth/keiji/pkg/controller"
	"github.com/gin-gonic/gin"
)

func Register(e *gin.Engine, root string, domain string, redisPort string, redisAddr string) {
	c := controller.NewController(root, domain, redisPort, redisAddr)
	web := e.Group("")
	web.GET("/", c.ServeBlogHome)
	web.GET("/home", c.ServeHome)
	web.GET("/blog", c.ServeBlogHome)
	web.GET("/creative", c.ServeCreativeWriting)
	web.GET("/technical", c.ServeTechnicalWriteups)
	web.GET("/digital", c.ServeDigitalArt)
	web.GET("/writing/:post-name", c.ServePost)
	web.GET("/login", c.ServeLogin)
	web.POST("/login", c.Auth)


	cdn := e.Group("/api/v1")
	cdn.GET("/style/:file", c.ServeCss)
	cdn.GET("/js/:file", c.ServeJs)
	cdn.GET("/style/mdb/:file", c.ServeMdbCss)
	cdn.GET("/assets/:file", c.ServeAsset)
	cdn.GET("/images/:file", c.ServeImage)
	cdn.GET("/cdn/:file", c.ServeGeneric)	cdn.GET("/htmx/:file", c.ServeHtmx)



	priv := e.Group("/admin")
	priv.Use(c.IsAuthenticated)
	priv.GET("/upload", c.ServeFileUpload)
	priv.POST("/upload", c.SaveFile)
	priv.GET("/panel", c.AdminPanel)
	priv.POST("/add-document", c.AddDocument)
	priv.POST("/images/upload", c.SaveFile)
	priv.GET("/posts/:post-name", c.GetBlogPostEditor)
	priv.POST("/posts", c.MakeBlogPost)
	priv.GET("/posts/all", c.ServeBlogDirectory)
	priv.GET("/posts", c.ServeNewBlogPage)
	priv.PATCH("/posts", c.UpdateBlogPost)

}
