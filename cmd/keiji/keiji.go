package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"strconv"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"

	"git.aetherial.dev/aeth/keiji/pkg/auth"
	"git.aetherial.dev/aeth/keiji/pkg/env"
	"git.aetherial.dev/aeth/keiji/pkg/routes"
	"git.aetherial.dev/aeth/keiji/pkg/storage"
	"git.aetherial.dev/aeth/keiji/pkg/webpages"

	_ "github.com/mattn/go-sqlite3"
)

var contentMode, envPath string
var blank bool

func printUsage() string {
	return "wrong"
}

func main() {
	flag.StringVar(&contentMode, "content", "", "pass the option to run the webserver using filesystem or embedded html")
	flag.StringVar(&envPath, "env", ".env", "pass specific ..env file to the program startup")
	flag.BoolVar(&blank, "blank", false, "create a blank .env template")
	flag.Parse()
	if blank {
		fh, err := os.OpenFile(".env.template", os.O_CREATE|os.O_RDWR, os.ModePerm)
		if err != nil {
			log.Fatal("Couldnt open file .env.template, error: ", err)
		}
		defer fh.Close()
		env.WriteTemplate(fh)
		fmt.Println("Blank template written to: .env.template")
		os.Exit(0)
	}
	err := env.LoadAndVerifyEnv(envPath, env.REQUIRED_VARS)
	if err != nil {
		log.Fatal("Error when loading env file: ", err)
	}
	var srcOpt webpages.ServiceOption
	switch contentMode {
	case "fs":
		srcOpt = webpages.FILESYSTEM
	case "embed":
		srcOpt = webpages.EMBED
	default:
		printUsage()
		os.Exit(1)
	}
	var htmlReader fs.FS
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
	e := gin.Default()
	if srcOpt == webpages.FILESYSTEM {
		e.LoadHTMLGlob(path.Join(os.Getenv("WEB_ROOT"), "html", "*.html"))
	} else {
		for i := range templateNames {
			name := templateNames[i]
			renderer.AddFromString(
				name,
				webpages.ReadToString(htmlReader, path.Join("html", name+".html")),
			)
		}
		e.HTMLRender = renderer
	}
	dbfile := "sqlite.db"
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatal(err)
	}
	webserverDb := storage.NewSQLiteRepo(db)
	err = webserverDb.Migrate()
	if err != nil {
		log.Fatal(err)
	}
	routes.Register(e, os.Getenv("DOMAIN_NAME"), webserverDb, htmlReader, auth.EnvAuth{})
	ssl, err := strconv.ParseBool(os.Getenv("USE_SSL"))
	if err != nil {
		log.Fatal("Invalid option passed to USE_SSL: ", os.Getenv("USE_SSL"))
	}
	if ssl {
		e.RunTLS(fmt.Sprintf("%s:%s", os.Getenv("HOST_ADDR"), os.Getenv("HOST_PORT")),
			os.Getenv(env.CHAIN), os.Getenv(env.KEY))
	}
	e.Run(fmt.Sprintf("%s:%s", os.Getenv("HOST_ADDR"), os.Getenv("HOST_PORT")))
}
