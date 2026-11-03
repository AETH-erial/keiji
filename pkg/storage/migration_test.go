package storage

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func migrationDB(t *testing.T, schema string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err = db.Exec(schema); err != nil {
		t.Fatal(err)
	}
	return db
}
func TestMigrateLegacyImages(t *testing.T) {
	db := migrationDB(t, `CREATE TABLE images (id TEXT, title TEXT)`)
	dir := t.TempDir()
	ids := []string{"00000000-0000-4000-8000-000000000001", "00000000-0000-4000-8000-000000000002", "00000000-0000-4000-8000-000000000003", "00000000-0000-4000-8000-000000000004", "../unsafe", "00000000-0000-4000-8000-000000000005"}
	for _, id := range ids {
		if _, err := db.Exec(`INSERT INTO images VALUES (?, 'title')`, id); err != nil {
			t.Fatal(err)
		}
	}
	files := [][]byte{[]byte("GIF89a\x00\x00\x00"), {0x1a, 0x45, 0xdf, 0xa3, 0x87, 0x42, 0x82, 0x84, 'w', 'e', 'b', 'm'}, {0x1a, 0x45, 0xdf, 0xa3, 0x8b, 0x42, 0x82, 0x88, 'm', 'a', 't', 'r', 'o', 's', 'k', 'a'}}
	for i, data := range files {
		if err := os.WriteFile(filepath.Join(dir, ids[i]), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join(dir, ids[0]), filepath.Join(dir, ids[5])); err != nil {
		t.Fatal(err)
	}
	report, err := MigrateLegacyImages(context.Background(), db, dir, "digital_art")
	if err != nil {
		t.Fatal(err)
	}
	if report != (MigrationReport{Examined: 6, Updated: 2, Unknown: 4}) {
		t.Fatalf("report: %+v", report)
	}
	for i, id := range ids {
		var category string
		var mt sql.NullString
		if err := db.QueryRow(`SELECT category, media_type FROM images WHERE id=?`, id).Scan(&category, &mt); err != nil {
			t.Fatal(err)
		}
		if category != "digital_art" {
			t.Fatal(category)
		}
		want := ""
		if i == 0 {
			want = "image/gif"
		}
		if i == 1 {
			want = "video/webm"
		}
		if mt.String != want || (want == "" && mt.Valid) {
			t.Fatalf("%s: %+v", id, mt)
		}
	}
	report, err = MigrateLegacyImages(context.Background(), db, dir, "other")
	if err != nil {
		t.Fatal(err)
	}
	if report != (MigrationReport{Examined: 4, Unknown: 4}) {
		t.Fatalf("second: %+v", report)
	}
}
func TestMigrateExistingColumns(t *testing.T) {
	db := migrationDB(t, `CREATE TABLE images (id TEXT, category TEXT, media_type TEXT NOT NULL DEFAULT '')`)
	id := "00000000-0000-4000-8000-000000000001"
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, id), []byte("GIF89a"), 0600)
	if _, err := db.Exec(`INSERT INTO images VALUES (?, 'old', ''), (?, 'keep', 'custom/type')`, id, id); err != nil {
		t.Fatal(err)
	}
	report, err := MigrateLegacyImages(context.Background(), db, dir, "new")
	if err != nil {
		t.Fatal(err)
	}
	if report.Updated != 1 {
		t.Fatal(report)
	}
	var cat, mt string
	if err := db.QueryRow(`SELECT category, media_type FROM images WHERE category='keep'`).Scan(&cat, &mt); err != nil || mt != "custom/type" {
		t.Fatalf("%s %s %v", cat, mt, err)
	}
}
func TestMigrationRollback(t *testing.T) {
	db := migrationDB(t, `CREATE TABLE images (id TEXT); INSERT INTO images VALUES ('00000000-0000-4000-8000-000000000001'); CREATE TRIGGER reject_update BEFORE UPDATE ON images BEGIN SELECT RAISE(ABORT,'reject'); END;`)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "00000000-0000-4000-8000-000000000001"), []byte("GIF89a"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := MigrateLegacyImages(context.Background(), db, dir, "digital_art"); err == nil {
		t.Fatal("expected error")
	}
	rows, err := db.Query(`PRAGMA table_info(images)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
	}
	if count != 1 {
		t.Fatalf("DDL not rolled back: %d columns", count)
	}
}
