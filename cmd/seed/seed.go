package main

import (
	"log"
	"os"
	"database/sql"

	"git.aetherial.dev/aeth/keiji/pkg/env"
	"git.aetherial.dev/aeth/keiji/pkg/helpers"
	_ "github.com/mattn/go-sqlite3"
)


func main() {
	err := env.LoadAndVerifyEnv(os.Args[1], env.REQUIRED_VARS); if err != nil {
		log.Fatal(err)
	}
	dbfile := "sqlite.db"
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatal(err)
	}
	webserverDb := helpers.NewSQLiteRepo(db)
	err = webserverDb.Migrate()
	if err != nil {
		log.Fatal(err)
	} 
	webserverDb.Seed(os.Args[2], os.Args[3], os.Args[4])
	
}
