package controller

import (
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.aetherial.dev/aeth/keiji/pkg/storage"
	"github.com/gin-gonic/gin"
)

type htmlLegacyRepo struct {
	storage.DocumentIO
	category string
}

func (r *htmlLegacyRepo) GetDocument(storage.Identifier) (storage.Document, error) {
	return storage.Document{Title: "post"}, nil
}
func (r *htmlLegacyRepo) GetByCategory(category string) []storage.Document {
	r.category = category
	return nil
}
func (r *htmlLegacyRepo) GetImagesByCategory(category string) []storage.Image {
	r.category = category
	return nil
}
func (r *htmlLegacyRepo) GetNavBarLinks() []storage.NavBarItem    { return nil }
func (r *htmlLegacyRepo) GetDropdownElements() []storage.LinkPair { return nil }

type htmlErrorRepo struct {
	*htmlLegacyRepo
	storage.DocumentListReader
	fail string
}

func (r *htmlErrorRepo) listError(name string) error {
	if r.fail == name {
		return errors.New("list failure: " + name)
	}
	return nil
}
func (r *htmlErrorRepo) GetByCategoryWithError(category string) ([]storage.Document, error) {
	r.category = category
	return nil, r.listError("documents")
}
func (r *htmlErrorRepo) GetImagesByCategoryWithError(category string) ([]storage.Image, error) {
	r.category = category
	return nil, r.listError("images")
}
func (r *htmlErrorRepo) GetNavBarLinksWithError() ([]storage.NavBarItem, error) {
	return nil, r.listError("navbar")
}
func (r *htmlErrorRepo) GetDropdownElementsWithError() ([]storage.LinkPair, error) {
	return nil, r.listError("menu")
}

func htmlTestRouter(repo storage.DocumentIO) *gin.Engine {
	r := gin.New()
	r.SetHTMLTemplate(template.Must(template.New("test").Parse(`{{define "home"}}{{.default.Body}}{{end}}{{define "writing"}}writing{{end}}{{define "digital_art"}}digital{{end}}{{define "blogpost"}}{{.Title}}{{end}}`)))
	c := &Controller{database: repo}
	r.GET("/", c.ServeHome)
	r.GET("/blog", c.ServeBlog)
	r.GET("/creative", c.ServeCreative)
	r.GET("/digital", c.ServeDigitalArt)
	r.GET("/writing/:id", c.ServePost)
	return r
}

func TestHTMLListFailures(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct{ path, failure string }{
		{"/", "documents"}, {"/", "navbar"}, {"/", "menu"},
		{"/blog", "documents"}, {"/creative", "documents"},
		{"/digital", "images"}, {"/digital", "navbar"}, {"/digital", "menu"},
		{"/writing/id", "navbar"}, {"/writing/id", "menu"},
	} {
		t.Run(tc.path+"/"+tc.failure, func(t *testing.T) {
			repo := &htmlErrorRepo{htmlLegacyRepo: &htmlLegacyRepo{}, fail: tc.failure}
			response := httptest.NewRecorder()
			htmlTestRouter(repo).ServeHTTP(response, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if response.Code != http.StatusInternalServerError || !strings.Contains(response.Body.String(), "Could not load content") {
				t.Fatalf("got status %d body %q", response.Code, response.Body.String())
			}
		})
	}
}

func TestHTMLListEmptyAndLegacyCompatibility(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, legacy := range []bool{false, true} {
		for _, tc := range []struct{ path, category, body string }{
			{"/", storage.HOMEPAGE, "Under construction. Sry :("},
			{"/blog", storage.BLOG, "writing"}, {"/creative", storage.CREATIVE, "writing"},
			{"/digital", storage.DIGITAL_ART, "digital"}, {"/writing/id", "", "post"},
		} {
			t.Run(tc.path+map[bool]string{true: "/legacy", false: "/error-reader"}[legacy], func(t *testing.T) {
				base := &htmlLegacyRepo{}
				var repo storage.DocumentIO = base
				if !legacy {
					repo = &htmlErrorRepo{htmlLegacyRepo: base}
				}
				response := httptest.NewRecorder()
				htmlTestRouter(repo).ServeHTTP(response, httptest.NewRequest(http.MethodGet, tc.path, nil))
				if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), tc.body) || base.category != tc.category {
					t.Fatalf("got status %d body %q category %q", response.Code, response.Body.String(), base.category)
				}
			})
		}
	}
}

func (r *htmlErrorRepo) GetAssetsWithError() ([]storage.Asset, error) {
	return nil, r.listError("assets")
}
func (r *htmlLegacyRepo) GetAssets() []storage.Asset { return nil }
func TestAssetListFailureReturns500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := &htmlErrorRepo{htmlLegacyRepo: &htmlLegacyRepo{}, fail: "assets"}
	c := &Controller{database: repo}
	r.GET("/assets/:file", c.ServeAsset)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/assets/test.png", nil))
	if rec.Code != http.StatusInternalServerError || !strings.Contains(rec.Body.String(), "Could not load content") {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
}
