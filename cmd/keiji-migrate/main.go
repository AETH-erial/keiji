package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"git.aetherial.dev/aeth/keiji/pkg/storage"
	_ "github.com/mattn/go-sqlite3"
)

func run(args []string) error {
	flags := flag.NewFlagSet("keiji-migrate", flag.ContinueOnError)
	dbPath := flags.String("db", "", "existing SQLite database file")
	media := flags.String("media-dir", "", "directory containing UUID media files")
	category := flags.String("legacy-category", "", "category for all rows if category column is missing")
	offline := flags.Bool("offline", false, "confirm application and other writers are stopped")
	backup := flags.Bool("backup-confirmed", false, "confirm a verified backup exists")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *dbPath == "" || *media == "" || *category == "" || !*offline || !*backup {
		return fmt.Errorf("required: -db, -media-dir, -legacy-category, -offline, and -backup-confirmed")
	}
	abs, err := filepath.Abs(*dbPath)
	if err != nil {
		return err
	}
	uri := (&url.URL{Scheme: "file", Path: abs, RawQuery: "mode=rw&_busy_timeout=5000&_txlock=immediate"}).String()
	db, err := sql.Open("sqlite3", uri)
	if err != nil {
		return err
	}
	defer db.Close()
	report, err := storage.MigrateLegacyImages(context.Background(), db, *media, *category)
	if err != nil {
		return err
	}
	fmt.Printf("Migration committed: examined=%d updated=%d unknown=%d\n", report.Examined, report.Updated, report.Unknown)
	return nil
}
func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
