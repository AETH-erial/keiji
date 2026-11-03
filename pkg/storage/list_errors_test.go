package storage

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// Every list operation must report query and scan failures without terminating
// the process or leaking the single database connection.
func TestListOperationsReportErrors(t *testing.T) {
	tests := []struct {
		name, table, schema string
		call                func(*SQLiteRepo) error
	}{
		{"menu", "menu", "id INTEGER, link TEXT, text TEXT, extra TEXT", func(r *SQLiteRepo) error { _, e := r.GetDropdownElementsWithError(); return e }},
		{"navbar", "navbar", "id INTEGER, png BLOB, link TEXT, redirect TEXT, extra TEXT", func(r *SQLiteRepo) error { _, e := r.GetNavBarLinksWithError(); return e }},
		{"assets", "assets", "id INTEGER, name TEXT, data BLOB, extra TEXT", func(r *SQLiteRepo) error { _, e := r.GetAssetsWithError(); return e }},
		{"admin", "admin", "id INTEGER, display_name TEXT, link TEXT, category TEXT, extra TEXT", func(r *SQLiteRepo) error { _, e := r.GetAdminTablesWithError(); return e }},
		{"category", "posts", "row INTEGER, id TEXT, title TEXT, created TEXT, body TEXT, category TEXT, sample TEXT, extra TEXT", func(r *SQLiteRepo) error { _, e := r.GetByCategoryWithError("test"); return e }},
		{"documents", "posts", "row INTEGER, id TEXT, title TEXT, created TEXT, body TEXT, category TEXT, sample TEXT, extra TEXT", func(r *SQLiteRepo) error { _, e := r.AllDocumentsWithError(); return e }},
		// NULL title causes Scan to fail after a successful query.
		{"images", "images", "id TEXT, title TEXT, desc TEXT, created TEXT, category TEXT, media_type TEXT", func(r *SQLiteRepo) error { _, e := r.GetAllImagesWithError(); return e }},
		{"images category", "images", "id TEXT, title TEXT, desc TEXT, created TEXT, category TEXT, media_type TEXT", func(r *SQLiteRepo) error { _, e := r.GetImagesByCategoryWithError("test"); return e }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := sql.Open("sqlite3", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			db.SetMaxOpenConns(1)
			repo := NewSQLiteRepo(db, nil)
			if err := tt.call(repo); err == nil {
				t.Fatal("missing table: expected query error")
			}
			if _, err := db.Exec("CREATE TABLE " + tt.table + " (" + tt.schema + ")"); err != nil {
				t.Fatal(err)
			}
			if tt.table == "posts" || tt.table == "images" {
				_, err = db.Exec("INSERT INTO " + tt.table + " (category) VALUES ('test')")
			} else {
				_, err = db.Exec("INSERT INTO " + tt.table + " DEFAULT VALUES")
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := tt.call(repo); err == nil {
				t.Fatal("bad row: expected scan error")
			}
			if stats := db.Stats(); stats.InUse != 0 {
				t.Fatalf("connection still in use: %+v", stats)
			}
		})
	}
}

func TestLegacyListMethodsDoNotExitOnQueryError(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewSQLiteRepo(db, nil)
	repo.GetDropdownElements()
	repo.GetNavBarLinks()
	repo.GetAssets()
	repo.GetAdminTables()
	repo.GetByCategory("test")
	repo.AllDocuments()
	repo.GetAllImages()
	repo.GetImagesByCategory("test")
}
