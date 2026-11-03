package main

import (
	"database/sql"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestRun(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "with space?#.db")
	db, err := sql.Open("sqlite3", (&url.URL{Scheme: "file", Path: dbPath}).String())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`CREATE TABLE images (id TEXT)`); err != nil {
		t.Fatal(err)
	}
	db.Close()
	args := []string{"-db", dbPath, "-media-dir", dir, "-legacy-category", "digital_art", "-offline", "-backup-confirmed"}
	if err = run(args[:len(args)-1]); err == nil {
		t.Fatal("missing acknowledgement accepted")
	}
	if err = run(args); err != nil {
		t.Fatal(err)
	}
	args[1] = filepath.Join(dir, "missing.db")
	if err = run(args); err == nil {
		t.Fatal("missing database accepted")
	}
	if _, err = os.Stat(args[1]); !os.IsNotExist(err) {
		t.Fatal("created missing database")
	}
}
