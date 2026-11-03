package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"git.aetherial.dev/aeth/keiji/pkg/controller"
	"github.com/gin-gonic/gin"
)

func TestNewRequestOrigin(t *testing.T) {
	for _, tc := range []struct{ target, origin string }{
		{"https://example.com:8443/a?x=1#fragment", "https://example.com:8443"},
		{"http://[::1]:8080/path", "http://[::1]:8080"},
	} {
		req, err := newRequest(http.MethodPost, tc.target, nil)
		if err != nil || req.Header.Get("Origin") != tc.origin {
			t.Errorf("%s: request=%v err=%v", tc.target, req, err)
		}
	}
	for _, target := range []string{"/relative", "ftp://example.com/path", "https://user:pass@example.com", "http:///path", "://invalid"} {
		if _, err := newRequest(http.MethodPost, target, nil); err == nil {
			t.Errorf("accepted invalid target %q", target)
		}
	}
	req, err := newRequest(http.MethodGet, "https://example.com", nil)
	if err != nil || req.Header.Get("Origin") != "" {
		t.Fatalf("GET: request=%v err=%v", req, err)
	}
	if _, err := newfileUploadRequest("://invalid", nil, "file", createFile(t)); err == nil {
		t.Fatal("upload accepted invalid target")
	}
}

func createFile(t *testing.T) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "image.png")
	if err := os.WriteFile(file, []byte("image"), 0600); err != nil {
		t.Fatal(err)
	}
	return file
}

// Run the real CLI in a subprocess because several command paths call os.Exit.
func TestCLIHelper(t *testing.T) {
	if os.Getenv("KEIJI_CTL_TEST_HELPER") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{"keiji-ctl"}, os.Args[i+1:]...)
			flag.CommandLine = flag.NewFlagSet("keiji-ctl", flag.ExitOnError)
			main()
			return
		}
	}
	t.Fatal("missing helper arguments")
}

func TestCLIWithSameOrigin(t *testing.T) {
	t.Setenv("KEIJI_INSECURE_COOKIE", "true")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/login", controller.SameOrigin, func(c *gin.Context) {
		http.SetCookie(c.Writer, &http.Cookie{Name: controller.AUTH_COOKIE_NAME, Value: "test-cookie"})
		c.Status(http.StatusOK)
	})
	for _, endpoint := range []string{"asset", "navbar", "menu", "panel"} {
		r.POST("/admin/"+endpoint, controller.SameOrigin, func(c *gin.Context) {
			cookie, err := c.Request.Cookie(controller.AUTH_COOKIE_NAME)
			if err != nil || cookie.Value != "test-cookie" {
				c.Status(http.StatusUnauthorized)
				return
			}
			c.Status(http.StatusOK)
		})
	}
	server := httptest.NewServer(r)
	defer server.Close()
	file := createFile(t)
	for _, command := range []string{"auth", "asset", "nav", "menu", "admin"} {
		t.Run(command, func(t *testing.T) {
			args := []string{"-test.run=^TestCLIHelper$", "--", "-cmd", command, "-address", server.URL, "-cookie", "test-cookie", "-png", file}
			cmd := exec.Command(os.Args[0], args...)
			cmd.Env = append(os.Environ(), "KEIJI_CTL_TEST_HELPER=1")
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("CLI failed: %v: %s", err, output)
			}
			if command == "auth" && !strings.Contains(string(output), "test-cookie") {
				t.Fatalf("missing auth cookie: %s", output)
			}
		})
	}
}

func TestRedirectNotFollowed(t *testing.T) {
	for _, code := range []int{http.StatusFound, http.StatusTemporaryRedirect, http.StatusPermanentRedirect} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			called := false
			destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))
			defer destination.Close()
			source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, destination.URL, code)
			}))
			defer source.Close()
			req, err := newRequest(http.MethodPost, source.URL+"/login", strings.NewReader("credentials"))
			if err != nil {
				t.Fatal(err)
			}
			req.AddCookie(&http.Cookie{Name: "session", Value: "secret"})
			resp, err := newHTTPClient().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != code || called {
				t.Fatalf("status=%d, destination called=%v", resp.StatusCode, called)
			}
		})
	}
}
