package controller

import (
	"bytes"
	"database/sql"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"git.aetherial.dev/aeth/keiji/pkg/storage"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// Run explicitly with KEIJI_BACKUP_DIR pointing to a backup containing the
// migrated database and image store. The database is opened read-only.
func TestMigratedBackup(t *testing.T) {
	dir := os.Getenv("KEIJI_BACKUP_DIR")
	if dir == "" {
		t.Skip("set KEIJI_BACKUP_DIR to test a migrated backup")
	}
	db, err := sql.Open("sqlite3", "file:"+filepath.Join(dir, "sqlite-media-types-category.db")+"?mode=ro&immutable=1")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := storage.NewSQLiteRepo(db, storage.FilesystemImageIO{RootDir: filepath.Join(dir, "images")})
	images := repo.GetImagesByCategory(storage.DIGITAL_ART)
	if len(images) != 8 {
		t.Fatalf("got %d images, want 8", len(images))
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	c := &Controller{database: repo}
	r.GET("/images/:file", c.ServeImage)
	counts := map[string]int{}
	for _, img := range images {
		counts[img.MediaType]++
		original, err := os.ReadFile(filepath.Join(dir, "images", string(img.Ident)))
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest("GET", "/images/"+string(img.Ident), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != 200 || rec.Header().Get("Content-Type") != img.MediaType || !bytes.Equal(rec.Body.Bytes(), original) {
			t.Fatalf("bad response for %s: status %d type %s", img.Ident, rec.Code, rec.Header().Get("Content-Type"))
		}
		if img.MediaType == "video/webm" {
			if detectMediaType(original) != "video/webm" {
				t.Fatal("webm detection failed")
			}
			req = httptest.NewRequest("GET", "/images/"+string(img.Ident), nil)
			req.Header.Set("Range", "bytes=0-15")
			rec = httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != 206 || !bytes.Equal(rec.Body.Bytes(), original[:16]) {
				t.Fatal("webm range request failed")
			}
		}
	}
	if counts["video/webm"] != 1 || counts["image/jpeg"] != 3 || counts["image/gif"] != 4 {
		t.Fatalf("unexpected types: %v", counts)
	}
}
