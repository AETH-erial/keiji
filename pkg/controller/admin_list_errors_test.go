package controller

import (
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.aetherial.dev/aeth/keiji/pkg/auth"
	"git.aetherial.dev/aeth/keiji/pkg/storage"
	"github.com/gin-gonic/gin"
)

type adminLegacyListRepo struct{ *htmlLegacyRepo }

func (r *adminLegacyListRepo) GetAdminTables() storage.AdminPage { return storage.AdminPage{} }
func (r *adminLegacyListRepo) GetAllImages() []storage.Image     { return nil }

type adminErrorListRepo struct {
	*adminLegacyListRepo
	storage.DocumentListReader
	fail string
}

func (r *adminErrorListRepo) failure(name string) error {
	if r.fail == name {
		return errors.New("list failure: " + name)
	}
	return nil
}
func (r *adminErrorListRepo) GetAdminTablesWithError() (storage.AdminPage, error) {
	return storage.AdminPage{}, r.failure("tables")
}
func (r *adminErrorListRepo) GetAllImagesWithError() ([]storage.Image, error) {
	return nil, r.failure("all images")
}
func (r *adminErrorListRepo) GetImagesByCategoryWithError(category string) ([]storage.Image, error) {
	r.category = category
	return nil, r.failure("images")
}
func (r *adminErrorListRepo) GetByCategoryWithError(category string) ([]storage.Document, error) {
	r.category = category
	return nil, r.failure("documents")
}
func (r *adminErrorListRepo) GetNavBarLinksWithError() ([]storage.NavBarItem, error) {
	return nil, r.failure("navbar")
}
func (r *adminErrorListRepo) GetDropdownElementsWithError() ([]storage.LinkPair, error) {
	return nil, r.failure("menu")
}

type adminListAuth struct{}

func (adminListAuth) AdminUsername() string { return "user" }
func (adminListAuth) AdminPassword() string { return "pass" }

func adminListRouter(repo storage.DocumentIO) *gin.Engine {
	r := gin.New()
	r.SetHTMLTemplate(template.Must(template.New("test").Parse(`{{define "admin"}}admin {{range index .Tables "modify"}}{{.DisplayName}}{{end}}{{end}}{{define "blogpost_editor"}}editor{{end}}{{define "navbar_editor"}}navbar editor{{end}}{{define "upload"}}upload{{end}}{{define "digital_art_admin_list"}}digital list{{end}}`)))
	c := &Controller{database: repo, Cache: auth.NewCache(), AuthSource: adminListAuth{}}
	r.POST("/login", c.Auth)
	r.GET("/panel", c.AdminPanel)
	r.GET("/navbar", c.ServeNavbarModify)
	r.GET("/blog", c.ServeBlogDirectory)
	r.GET("/editor/:id", c.GetBlogPostEditor)
	r.GET("/new-blog", c.ServeNewBlogPage)
	r.GET("/new-navbar", c.ServeNewNavbarItem)
	r.GET("/upload", c.ServeFileUpload)
	r.GET("/digital", c.ListDigitalArt)
	return r
}

func adminListRequest(r *gin.Engine, path string) *httptest.ResponseRecorder {
	method, body := http.MethodGet, ""
	if path == "/login" {
		method, body = http.MethodPost, "username=user&password=pass"
	}
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	response := httptest.NewRecorder()
	r.ServeHTTP(response, req)
	return response
}

func TestAdminListFailures(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct{ path, failure string }{
		{"/login", "tables"}, {"/login", "navbar"}, {"/login", "menu"},
		{"/panel", "tables"}, {"/panel", "navbar"}, {"/panel", "menu"},
		{"/navbar", "navbar"}, {"/navbar", "menu"},
		{"/blog", "documents"}, {"/blog", "navbar"}, {"/blog", "menu"},
		{"/editor/id", "all images"}, {"/editor/id", "navbar"}, {"/editor/id", "menu"},
		{"/new-blog", "all images"}, {"/new-blog", "navbar"}, {"/new-blog", "menu"},
		{"/new-navbar", "navbar"}, {"/new-navbar", "menu"},
		{"/upload", "navbar"}, {"/upload", "menu"}, {"/digital", "images"},
	} {
		t.Run(tc.path+"/"+tc.failure, func(t *testing.T) {
			repo := &adminErrorListRepo{adminLegacyListRepo: &adminLegacyListRepo{&htmlLegacyRepo{}}, fail: tc.failure}
			response := adminListRequest(adminListRouter(repo), tc.path)
			if response.Code != http.StatusInternalServerError || !strings.Contains(response.Body.String(), "Could not load content") {
				t.Fatalf("got status %d body %q", response.Code, response.Body.String())
			}
			if tc.path == "/login" && len(response.Result().Cookies()) != 0 {
				t.Fatal("failed login response set a cookie")
			}
		})
	}
}

func TestAdminListEmptyAndLegacyCompatibility(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, legacy := range []bool{false, true} {
		for _, tc := range []struct{ path, body string }{
			{"/login", "Digital Art"}, {"/panel", "Digital Art"}, {"/navbar", "admin"}, {"/blog", "admin"},
			{"/editor/id", "editor"}, {"/new-blog", "editor"}, {"/new-navbar", "navbar editor"}, {"/upload", "upload"}, {"/digital", "digital list"},
		} {
			t.Run(tc.path+map[bool]string{true: "/legacy", false: "/error-reader"}[legacy], func(t *testing.T) {
				base := &adminLegacyListRepo{&htmlLegacyRepo{}}
				var repo storage.DocumentIO = base
				if !legacy {
					repo = &adminErrorListRepo{adminLegacyListRepo: base}
				}
				response := adminListRequest(adminListRouter(repo), tc.path)
				if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), tc.body) {
					t.Fatalf("got status %d body %q", response.Code, response.Body.String())
				}
				if tc.path == "/digital" && base.category != storage.DIGITAL_ART {
					t.Fatalf("category = %q", base.category)
				}
				if tc.path == "/login" && len(response.Result().Cookies()) == 0 {
					t.Fatal("successful login did not set a cookie")
				}
			})
		}
	}
}
