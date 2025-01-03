package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"

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
	embedPtr := flag.Bool("embed", false, "Force the server to serve embedded content, for production use")
	fsPtr := flag.Bool("fs", false, "Force the server to serve embedded content, for production use")
	envPtr := flag.String("env", ".env", "pass specific ..env file to the program startup")
	flag.Parse()
	err := env.LoadAndVerifyEnv(*envPtr, env.REQUIRED_VARS)
	if err != nil {
		log.Fatal("Error when loading env file: ", err)
	}
	var srcOpt webpages.ServiceOption
	var htmlReader fs.FS
	if *embedPtr == true {
		srcOpt = webpages.EMBED
	}
	if *fsPtr == true {
		srcOpt = webpages.FILESYSTEM
	}
	htmlReader = webpages.NewContentLayer(srcOpt)
	renderer := multitemplate.NewDynamic()
	templateNames := []string{
		"home",
		"blogpost",
		"digital_art",
		"login",
		"admin",
		"blogpost_editor",
		"post_options",
		"unhandled_error",
		"upload",
		"upload_status",
		"writing",
		"listing",
	}
	if srcOpt == webpages.FILESYSTEM {
		for i := range templateNames {
			name := templateNames[i]
			filePath := path.Join(os.Getenv("WEB_ROOT"), "html", fmt.Sprintf("%s.html", name))
			fmt.Println(filePath)
			renderer.AddFromFiles(name, filePath)
		}
	} else {
		for i := range templateNames {
			name := templateNames[i]
			renderer.AddFromString(
				name,
				webpages.ReadToString(htmlReader, path.Join("html", name+".html")),
			)
		}
	}
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
	routes.Register(e, DOMAIN_NAME, webserverDb, htmlReader)
	if os.Getenv("SSL_MODE") == "ON" {
		e.RunTLS(fmt.Sprintf("%s:%s", os.Getenv("HOST_ADDR"), os.Getenv("HOST_PORT")),
			os.Getenv(env.CHAIN), os.Getenv(env.KEY))
	}
	e.Run(fmt.Sprintf("%s:%s", os.Getenv("HOST_ADDR"), os.Getenv("HOST_PORT")))
}
