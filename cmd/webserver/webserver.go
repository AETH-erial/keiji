package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"

	"git.aetherial.dev/aeth/keiji/pkg/env"
	"git.aetherial.dev/aeth/keiji/pkg/helpers"
	"git.aetherial.dev/aeth/keiji/pkg/routes"
	"git.aetherial.dev/aeth/keiji/pkg/webpages"

	_ "github.com/mattn/go-sqlite3"
)

var DOMAIN_NAME string

func main() {
	htmlSrc := flag.String("html-src", "", "Force the server to serve embedded content, for production use")
	flag.Parse()
	args := os.Args
	err := env.LoadAndVerifyEnv(args[1], env.REQUIRED_VARS)
	if err != nil {
		log.Fatal("Error when loading env file: ", err)
	}
	REDIS_PORT := os.Getenv("REDIS_PORT")
	REDIS_ADDR := os.Getenv("REDIS_ADDR")
	var srcOpt webpages.ServiceOption
	if *htmlSrc == "filesystem" {
		srcOpt = webpages.FILESYSTEM
	}
	if *htmlSrc == "embed" {
		srcOpt = webpages.EMBED
	}
	fmt.Println(srcOpt, *htmlSrc)
	// htmlReader := webpages.NewContentLayer(webpages.ServiceOption(webpages.FILESYSTEM))
	htmlReader := webpages.FilesystemWebpages{Webroot: os.Getenv("WEB_ROOT")}
	renderer := multitemplate.NewDynamic()
	renderer.AddFromString(
		"head",
		webpages.ReadToString(htmlReader, "head.html"),
	)
	renderer.AddFromString(
		"navigation",
		webpages.ReadToString(htmlReader, "navigation.html"),
	)
	renderer.AddFromString(
		"home",
		webpages.ReadToString(htmlReader, "home.html"),
	)
	renderer.AddFromString(
		"blogpost",
		webpages.ReadToString(htmlReader, "blogpost.html"),
	)
	renderer.AddFromString(
		"digital_art",
		webpages.ReadToString(htmlReader, "digital_art.html"),
	)
	renderer.AddFromString(
		"login",
		webpages.ReadToString(htmlReader, "login.html"),
	)
	renderer.AddFromString(
		"admin",
		webpages.ReadToString(htmlReader, "admin.html"),
	)
	renderer.AddFromString(
		"blogpost_editor",
		webpages.ReadToString(htmlReader, "blogpost_editor.html"),
	)
	renderer.AddFromString(
		"new_blogpost",
		webpages.ReadToString(htmlReader, "new_blogpost.html"),
	)
	renderer.AddFromString(
		"upload_status",
		webpages.ReadToString(htmlReader, "upload_status.html"),
	)
	renderer.AddFromString(
		"unhandled_error",
		webpages.ReadToString(htmlReader, "unhandled_error.html"),
	)
	renderer.AddFromString(
		"upload",
		webpages.ReadToString(htmlReader, "upload.html"),
	)
	renderer.AddFromString(
		"writing",
		webpages.ReadToString(htmlReader, "writing.html"),
	)
	renderer.AddFromString(
		"listing",
		webpages.ReadToString(htmlReader, "listing.html"),
	)
	e := gin.Default()
	dbfile := "sqlite.db"
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatal(err)
	}
	e.HTMLRender = renderer
	// 	e.LoadHTMLGlob("pkg/webpages/html/*.html")
	webserverDb := helpers.NewSQLiteRepo(db)
	err = webserverDb.Migrate()
	if err != nil {
		log.Fatal(err)
	}
	routes.Register(e, DOMAIN_NAME, REDIS_PORT, REDIS_ADDR, webserverDb)
	if os.Getenv("SSL_MODE") == "ON" {
		e.RunTLS(fmt.Sprintf("%s:%s", os.Getenv("HOST_ADDR"), os.Getenv("HOST_PORT")),
			os.Getenv(env.CHAIN), os.Getenv(env.KEY))
	}
	e.Run(fmt.Sprintf("%s:%s", os.Getenv("HOST_ADDR"), os.Getenv("HOST_PORT")))
}
