package routes

import (
	"net/http/httptest"
	"testing"

	"git.aetherial.dev/aeth/keiji/pkg/auth"
	"git.aetherial.dev/aeth/keiji/pkg/storage"
	"git.aetherial.dev/aeth/keiji/pkg/webpages"
	"github.com/gin-gonic/gin"
)

func TestRegister(t *testing.T) {
	e := gin.Default()
	Register(e, "localhost", &storage.SQLiteRepo{}, webpages.FilesystemWebpages{}, auth.EnvAuth{})
}

func TestDigitalArtRoutesRequireAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	Register(e, "localhost", &storage.SQLiteRepo{}, webpages.FilesystemWebpages{}, auth.EnvAuth{})
	for _, tc := range []struct{ method, path string }{{"GET", "/admin/digital"}, {"GET", "/admin/digital/123"}, {"POST", "/admin/digital/123"}, {"DELETE", "/admin/digital/123"}} {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
		if rec.Code != 401 && (rec.Code < 300 || rec.Code >= 400) {
			t.Fatalf("%s %s: expected redirect or unauthorized, got %d", tc.method, tc.path, rec.Code)
		}
	}
}

func TestLoginSecurityMiddlewareWired(t *testing.T) {
	t.Setenv("KEIJI_INSECURE_COOKIE", "")
	t.Setenv("KEIJI_TRUSTED_PROXIES", "")
	gin.SetMode(gin.TestMode)
	e := gin.New()
	Register(e, "localhost", &storage.SQLiteRepo{}, webpages.FilesystemWebpages{}, auth.EnvAuth{})
	for i := 0; i < 11; i++ {
		req := httptest.NewRequest("POST", "https://example.com/login", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		want := 403
		if i == 10 {
			want = 429
		}
		if rec.Code != want {
			t.Fatalf("attempt %d: got %d, want %d", i+1, rec.Code, want)
		}
	}
}
