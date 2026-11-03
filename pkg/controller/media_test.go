package controller

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.aetherial.dev/aeth/keiji/pkg/storage"
	"github.com/gin-gonic/gin"
)

type mediaRepo struct {
	storage.DocumentIO
	image  storage.Image
	images []storage.Image
	err    error
	added  bool
}

func (r *mediaRepo) AddImage(data []byte, title, desc, category, mediaType string) (storage.Identifier, error) {
	r.added = true
	r.image = storage.Image{Data: data, Title: title, Desc: desc, Category: category, MediaType: mediaType}
	return "id", r.err
}
func (r *mediaRepo) GetImage(id storage.Identifier) (storage.Image, error) { return r.image, r.err }
func (r *mediaRepo) GetAllImages() []storage.Image                         { return r.images }
func (r *mediaRepo) GetDropdownElements() []storage.LinkPair               { return nil }
func (r *mediaRepo) GetNavBarLinks() []storage.NavBarItem                  { return nil }
func (r *mediaRepo) GetDocument(id storage.Identifier) (storage.Document, error) {
	return storage.Document{Ident: id}, r.err
}

func testRouter(repo *mediaRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.SetHTMLTemplate(template.Must(template.New("test").Parse(`{{define "upload_status"}}{{.UpdateMessage}}{{end}}{{define "blogpost_editor"}}{{range .Images}}{{.Ident}}:{{.MediaType}}:{{.Title}}{{end}}{{end}}`)))
	c := &Controller{database: repo}
	r.POST("/upload", c.SaveFile)
	r.GET("/images/:file", c.ServeImage)
	r.GET("/new", c.ServeNewBlogPage)
	r.GET("/edit/:id", c.GetBlogPostEditor)
	return r
}

func TestDetectMediaType(t *testing.T) {
	webm := []byte{0x1a, 0x45, 0xdf, 0xa3, 0x87, 0x42, 0x82, 0x84, 'w', 'e', 'b', 'm'}
	cases := []struct {
		name string
		data []byte
		want string
	}{
		{"jpeg", []byte{0xff, 0xd8, 0xff, 0xdb}, "image/jpeg"},
		{"png", []byte("\x89PNG\r\n\x1a\n"), "image/png"},
		{"gif", []byte("GIF89a"), "image/gif"},
		{"webp", []byte("RIFF\x04\x00\x00\x00WEBPVP8 "), "image/webp"},
		{"webm", webm, "video/webm"},
		{"matroska", []byte("\x1a\x45\xdf\xa3\x8b\x42\x82\x88matroska"), ""},
		{"fake webm", []byte("\x1a\x45\xdf\xa3\x87\x42\x83\x84webm"), ""},
		{"truncated", webm[:len(webm)-1], ""}, {"unknown", []byte("text"), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := detectMediaType(tc.data); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func upload(t *testing.T, r http.Handler, data []byte, filename string) *httptest.ResponseRecorder {
	t.Helper()
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for k, v := range map[string]string{"title": "title", "description": "desc", "category": "art"} {
		if err := w.WriteField(k, v); err != nil {
			t.Fatal(err)
		}
	}
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = part.Write(data); err != nil {
		t.Fatal(err)
	}
	w.Close()
	req := httptest.NewRequest("POST", "/upload", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	return res
}
func TestUpload(t *testing.T) {
	r := &mediaRepo{}
	router := testRouter(r)
	data := []byte("GIF89ahello")
	res := upload(t, router, data, "wrong.jpg")
	if res.Code != 200 || !r.added || r.image.MediaType != "image/gif" || !bytes.Equal(r.image.Data, data) || r.image.Title != "title" || r.image.Desc != "desc" || r.image.Category != "art" {
		t.Fatalf("status %d image %+v", res.Code, r.image)
	}
	r.added = false
	res = upload(t, router, []byte("not an image"), "image.png")
	if res.Code != 415 || r.added {
		t.Fatalf("unknown: %d", res.Code)
	}
	res = upload(t, router, make([]byte, maxUploadSize+1), "huge.png")
	if res.Code != 413 || r.added {
		t.Fatalf("large: %d", res.Code)
	}
	r.err = errors.New("db error")
	res = upload(t, router, data, "x.gif")
	if res.Code != 500 {
		t.Fatalf("db error: %d", res.Code)
	}
}
func TestServeImage(t *testing.T) {
	repo := &mediaRepo{image: storage.Image{Data: []byte("0123456789"), MediaType: "video/webm"}}
	router := testRouter(repo)
	req := httptest.NewRequest("GET", "/images/abc", nil)
	req.Header.Set("Range", "bytes=2-5")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != 206 || res.Body.String() != "2345" || res.Header().Get("Content-Type") != "video/webm" || res.Header().Get("Content-Range") != "bytes 2-5/10" {
		t.Fatalf("range: %d %s %v", res.Code, res.Body.String(), res.Header())
	}
	for _, tc := range []struct {
		err    error
		status int
	}{{storage.ErrNotExists, 404}, {errors.New("storage failure"), 500}} {
		repo.err = tc.err
		res = httptest.NewRecorder()
		router.ServeHTTP(res, httptest.NewRequest("GET", "/images/abc", nil))
		if res.Code != tc.status {
			t.Fatalf("%v: %d", tc.err, res.Code)
		}
	}
}
func TestEditorsIncludeImages(t *testing.T) {
	repo := &mediaRepo{images: []storage.Image{{Ident: "abc", MediaType: "video/webm", Title: "Movie"}}}
	router := testRouter(repo)
	for _, path := range []string{"/new", "/edit/post"} {
		t.Run(path, func(t *testing.T) {
			res := httptest.NewRecorder()
			router.ServeHTTP(res, httptest.NewRequest("GET", path, nil))
			if res.Code != 200 || !strings.Contains(res.Body.String(), fmt.Sprintf("%s:%s:%s", "abc", "video/webm", "Movie")) {
				t.Fatalf("%d %s", res.Code, res.Body.String())
			}
		})
	}
}

func TestDigitalArtAdminHandlers(t *testing.T) {
	repo := &digitalAdminRepo{mediaRepo: mediaRepo{image: storage.Image{Ident: "id", Category: storage.DIGITAL_ART, Title: "old"}}}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.SetHTMLTemplate(template.Must(template.New("test").Parse(`{{define "admin"}}{{range .Tables.modify}}{{.DisplayName}} {{.Link}}{{end}}{{end}}{{define "digital_art_admin_list"}}{{range .Images}}{{.Title}}{{end}}{{end}}{{define "digital_art_admin_editor"}}{{.Title}}{{end}}{{define "upload_status"}}{{.UpdateMessage}}{{end}}`)))
	c := &Controller{database: repo}
	r.GET("/panel", c.AdminPanel)
	r.GET("/digital", c.ListDigitalArt)
	r.GET("/digital/:id", c.EditDigitalArt)
	r.POST("/digital/:id", c.UpdateDigitalArt)
	for _, path := range []string{"/panel", "/digital", "/digital/id"} {
		res := httptest.NewRecorder()
		r.ServeHTTP(res, httptest.NewRequest("GET", path, nil))
		if res.Code != 200 {
			t.Fatalf("%s: %d", path, res.Code)
		}
		if path == "/panel" && !strings.Contains(res.Body.String(), "/admin/digital") {
			t.Fatal("admin entry missing")
		}
	}
	req := httptest.NewRequest("POST", "/digital/id", strings.NewReader("title=New+title&description=New+description"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != 200 || repo.title != "New title" || repo.desc != "New description" {
		t.Fatalf("update: %d %q %q", res.Code, repo.title, repo.desc)
	}
	req = httptest.NewRequest("POST", "/digital/id", strings.NewReader("title=+"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res = httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != 400 {
		t.Fatalf("blank title: %d", res.Code)
	}
	repo.image.Category = storage.MISC_ASSET
	res = httptest.NewRecorder()
	r.ServeHTTP(res, httptest.NewRequest("GET", "/digital/id", nil))
	if res.Code != 404 {
		t.Fatalf("non-digital: %d", res.Code)
	}
	repo.updateErr = storage.ErrNotExists
	req = httptest.NewRequest("POST", "/digital/id", strings.NewReader("title=Title"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res = httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != 404 {
		t.Fatalf("missing update: %d", res.Code)
	}
}

type digitalAdminRepo struct {
	mediaRepo
	title, desc string
	updateErr   error
}

func (r *digitalAdminRepo) GetAdminTables() storage.AdminPage {
	return storage.AdminPage{Tables: map[string][]storage.TableData{}}
}
func (r *digitalAdminRepo) GetImagesByCategory(category string) []storage.Image {
	if category == storage.DIGITAL_ART {
		return []storage.Image{r.image}
	}
	return nil
}
func (r *digitalAdminRepo) UpdateDigitalArt(id storage.Identifier, title, desc string) error {
	r.title = title
	r.desc = desc
	return r.updateErr
}

func TestRemoveDigitalArt(t *testing.T) {
	repo := &archiveTestRepo{}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.SetHTMLTemplate(template.Must(template.New("test").Parse(`{{define "upload_status"}}{{.UpdateMessage}}{{end}}`)))
	c := &Controller{database: repo}
	r.DELETE("/digital/:id", c.RemoveDigitalArt)
	for _, tc := range []struct {
		err     error
		status  int
		message string
	}{{nil, 200, "Listing removed"}, {storage.ErrNotExists, 404, "Digital art not found"}, {errors.New("db failure"), 500, "Could not remove"}} {
		repo.err = tc.err
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest("DELETE", "/digital/art123", nil))
		if rec.Code != tc.status || repo.id != "art123" || !strings.Contains(rec.Body.String(), tc.message) {
			t.Fatalf("got %d %q id=%q", rec.Code, rec.Body.String(), repo.id)
		}
	}
}

type archiveTestRepo struct {
	storage.DocumentIO
	id  storage.Identifier
	err error
}

func (r *archiveTestRepo) ArchiveDigitalArt(id storage.Identifier) error { r.id = id; return r.err }
