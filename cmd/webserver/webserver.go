package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"adeptus-mechanicus.void/git/keiji/docs"
	"adeptus-mechanicus.void/git/keiji/pkg/env"
	"adeptus-mechanicus.void/git/keiji/pkg/routes"
)

var WEB_ROOT string
var DOMAIN_NAME string
var REDIS_PORT string
var REDIS_ADDR string

func main() {
	args := os.Args
	err := env.LoadAndVerifyEnv(args[1], env.REQUIRED_VARS)
	if err != nil {
		log.Fatal("Error when loading env file: ", err)
	}
	docs.SwaggerInfo.Title = "Webserver Swagger documentation"
	docs.SwaggerInfo.Host = "127.0.0.1:8080"
	docs.SwaggerInfo.Schemes = []string{"http"}
	renderer := multitemplate.NewRenderer()
	renderer.AddFromFiles(
		"home",
		fmt.Sprintf("%s/templates/home.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/navigation.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/menu.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/link.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/listing.html", WEB_ROOT),
	)
	renderer.AddFromFiles(
		"blogpost",
		fmt.Sprintf("%s/templates/blogpost.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/navigation.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/menu.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/link.html", WEB_ROOT),

	)
	renderer.AddFromFiles(
		"digital_art",
		fmt.Sprintf("%s/templates/digital_art.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/centered_image.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/navigation.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/menu.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/link.html", WEB_ROOT),
	)
	renderer.AddFromFiles(
		"login",
		fmt.Sprintf("%s/templates/login.html", WEB_ROOT),
	)
	renderer.AddFromFiles(
		"admin",
		fmt.Sprintf("%s/templates/admin.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/menu.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/link.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/navigation.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/listing.html", WEB_ROOT),
	)
	renderer.AddFromFiles(
		"blogpost_editor",
		fmt.Sprintf("%s/templates/blogpost_editor.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/menu.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/link.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/navigation.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/listing.html", WEB_ROOT),
	)
	e := gin.Default()
	e.HTMLRender = renderer
	e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	routes.Register(e, WEB_ROOT, DOMAIN_NAME, REDIS_PORT, REDIS_ADDR)
	e.Run(fmt.Sprintf("%s:%s", os.Getenv("HOST_ADDR"), os.Getenv("HOST_PORT")))

}
