package main

import (
	"fmt"
	"log"
	"os"
	"database/sql"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"

	"git.aetherial.dev/aeth/keiji/pkg/env"
	"git.aetherial.dev/aeth/keiji/pkg/routes"
	"git.aetherial.dev/aeth/keiji/pkg/helpers"
)

var WEB_ROOT string
var DOMAIN_NAME string


func main() {
	args := os.Args
	err := env.LoadAndVerifyEnv(args[1], env.REQUIRED_VARS)
	if err != nil {
		log.Fatal("Error when loading env file: ", err)
	}
	REDIS_PORT := os.Getenv("REDIS_PORT")
	REDIS_ADDR := os.Getenv("REDIS_ADDR")
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
		fmt.Sprintf("%s/templates/blogpost_editor.html", WEB_ROOT),
	)
	renderer.AddFromFiles(
		"blogpost_editor",
		fmt.Sprintf("%s/templates/blogpost_editor.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/menu.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/link.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/upload_status.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/navigation.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/listing.html", WEB_ROOT),
	)
	renderer.AddFromFiles(
		"new_blogpost",
		fmt.Sprintf("%s/templates/new_blogpost.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/menu.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/link.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/upload_status.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/navigation.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/listing.html", WEB_ROOT),
	)
	renderer.AddFromFiles(
		"upload_status",
		fmt.Sprintf("%s/templates/upload_status.html", WEB_ROOT),
	)
	renderer.AddFromFiles(
		"unhandled_error",
		fmt.Sprintf("%s/templates/unhandled_error.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/menu.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/link.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/navigation.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/listing.html", WEB_ROOT),
	)
	renderer.AddFromFiles(
		"upload",
		fmt.Sprintf("%s/templates/upload.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/menu.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/link.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/navigation.html", WEB_ROOT),
		fmt.Sprintf("%s/templates/listing.html", WEB_ROOT),
	)
	e := gin.Default()
	dbfile := "sqlite.db"
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatal(err)
	}
	e.HTMLRender = renderer
	webserverDb := helpers.NewSQLiteRepo(db)
	err = webserverDb.Migrate()
	if err != nil {
		log.Fatal(err)
	}
	routes.Register(e, WEB_ROOT, DOMAIN_NAME, REDIS_PORT, REDIS_ADDR, webserverDb)
	if os.Getenv("SSL_MODE") == "ON" {
		e.RunTLS(fmt.Sprintf("%s:%s", os.Getenv("HOST_ADDR"), os.Getenv("HOST_PORT")),
		os.Getenv(env.CHAIN), os.Getenv(env.KEY))
	}
	e.Run(fmt.Sprintf("%s:%s", os.Getenv("HOST_ADDR"), os.Getenv("HOST_PORT")))
}
