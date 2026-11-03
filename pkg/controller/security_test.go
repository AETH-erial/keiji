package controller

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.aetherial.dev/aeth/keiji/pkg/auth"
	"git.aetherial.dev/aeth/keiji/pkg/storage"
	"github.com/gin-gonic/gin"
)

type securitySource struct{}

func (securitySource) AdminUsername() string { return "admin" }
func (securitySource) AdminPassword() string { return "password" }

type securityRepo struct{ storage.DocumentIO }

func (securityRepo) GetAdminTables() storage.AdminPage       { return storage.AdminPage{} }
func (securityRepo) GetNavBarLinks() []storage.NavBarItem    { return nil }
func (securityRepo) GetDropdownElements() []storage.LinkPair { return nil }

func TestAuthCookieSecurity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, insecure := range []bool{false, true} {
		t.Run(map[bool]string{false: "production", true: "development"}[insecure], func(t *testing.T) {
			if insecure {
				t.Setenv("KEIJI_INSECURE_COOKIE", "true")
			} else {
				t.Setenv("KEIJI_INSECURE_COOKIE", "")
			}
			c := NewController("example.com", securityRepo{}, nil, securitySource{})
			r := gin.New()
			r.SetHTMLTemplate(template.Must(template.New("admin").Parse(`{{define "admin"}}ok{{end}}`)))
			r.POST("/login", c.Auth)
			req := httptest.NewRequest("POST", "/login", strings.NewReader(`{"username":"admin","password":"password"}`))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			cookies := rec.Result().Cookies()
			if rec.Code != 200 || len(cookies) != 1 {
				t.Fatalf("status %d cookies %v", rec.Code, cookies)
			}
			cookie := cookies[0]
			if cookie.Name != AUTH_COOKIE_NAME || cookie.Domain != "" || cookie.Path != "/" || cookie.Secure == insecure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || cookie.MaxAge != 1200 || !c.Cache.Read(cookie.Value) {
				t.Fatalf("bad cookie: %+v", cookie)
			}
		})
	}
}

func TestSameOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("KEIJI_INSECURE_COOKIE", "")
	r := gin.New()
	r.Any("/", SameOrigin, func(c *gin.Context) { c.Status(204) })
	for _, tc := range []struct {
		method, origin, referer string
		want                    int
	}{
		{"GET", "", "", 204}, {"POST", "https://example.com", "", 204},
		{"PATCH", "", "https://example.com/edit?q=x", 204}, {"DELETE", "https://evil.com", "https://example.com/", 403},
		{"POST", "null", "", 403}, {"POST", "", "", 403}, {"POST", "http://example.com", "", 403},
		{"POST", "https://example.com.evil.com", "", 403}, {"POST", "https://example.com/path", "", 403},
		{"POST", "https://example.com:444", "", 403},
	} {
		req := httptest.NewRequest(tc.method, "https://example.com/", nil)
		req.Header.Set("Origin", tc.origin)
		req.Header.Set("Referer", tc.referer)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Errorf("%s %q %q: %d want %d", tc.method, tc.origin, tc.referer, rec.Code, tc.want)
		}
	}
	t.Setenv("KEIJI_INSECURE_COOKIE", "true")
	req := httptest.NewRequest("POST", "http://example.com/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != 204 {
		t.Fatal(rec.Code)
	}
}

func TestLoginRateLimitProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, trusted := range []bool{false, true} {
		t.Run(map[bool]string{false: "untrusted", true: "trusted"}[trusted], func(t *testing.T) {
			if trusted {
				t.Setenv("KEIJI_TRUSTED_PROXIES", "10.0.0.0/8")
			} else {
				t.Setenv("KEIJI_TRUSTED_PROXIES", "")
			}
			r := gin.New()
			r.POST("/", LoginRateLimit(), func(c *gin.Context) { c.Status(204) })
			request := func(remote, xff string) int {
				req := httptest.NewRequest("POST", "/", nil)
				req.RemoteAddr = remote
				req.Header.Set("X-Forwarded-For", xff)
				rec := httptest.NewRecorder()
				r.ServeHTTP(rec, req)
				return rec.Code
			}
			for i := 0; i < 10; i++ {
				if got := request("10.0.0.1:123", "1.1.1.1"); got != 204 {
					t.Fatal(got)
				}
			}
			if got := request("10.0.0.1:123", "1.1.1.1"); got != 429 {
				t.Fatal(got)
			}
			want := 429
			if trusted {
				want = 204
			}
			if got := request("10.0.0.1:123", "2.2.2.2"); got != want {
				t.Fatalf("got %d want %d", got, want)
			}
			if trusted {
				if got := request("10.0.0.1:123", "2.2.2.2, 1.1.1.1"); got != 429 {
					t.Fatalf("spoofed chain: %d", got)
				}
			}
		})
	}
}

func TestAuthenticationStopsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c := &Controller{Cache: auth.NewCache()}
	r := gin.New()
	r.GET("/admin", c.IsAuthenticated, func(c *gin.Context) { c.Status(204) })
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/admin", nil))
	if rec.Code != 302 || rec.Header().Get("Location") != "/login" {
		t.Fatalf("%d %v", rec.Code, rec.Header())
	}
}

func TestAuthenticatedAdminRequiresOriginForMutation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("KEIJI_INSECURE_COOKIE", "")
	c := &Controller{Cache: auth.NewCache()}
	session, err := auth.Authorize(&auth.Credentials{Username: "admin", Password: "password"}, c.Cache, securitySource{})
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	group := r.Group("/admin", c.IsAuthenticated, SameOrigin)
	group.GET("/test", func(c *gin.Context) { c.Status(204) })
	group.POST("/test", func(c *gin.Context) { c.Status(204) })
	group.PATCH("/test", func(c *gin.Context) { c.Status(204) })
	group.DELETE("/test", func(c *gin.Context) { c.Status(204) })
	for _, tc := range []struct {
		method, origin string
		want           int
	}{{"GET", "", 204}, {"POST", "", 403}, {"POST", "https://evil.example", 403}, {"POST", "https://example.com", 204}, {"PATCH", "https://example.com", 204}, {"DELETE", "https://example.com", 204}} {
		req := httptest.NewRequest(tc.method, "https://example.com/admin/test", nil)
		req.AddCookie(&http.Cookie{Name: AUTH_COOKIE_NAME, Value: session})
		req.Header.Set("Origin", tc.origin)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Errorf("%s origin %q: got %d want %d", tc.method, tc.origin, rec.Code, tc.want)
		}
	}
}
