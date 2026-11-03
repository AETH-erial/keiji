package controller

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func (c *Controller) IsAuthenticated(ctx *gin.Context) {
	cookie, err := ctx.Cookie(AUTH_COOKIE_NAME)
	if err != nil || !c.Cache.Read(cookie) {
		ctx.Redirect(http.StatusFound, "/login")
		ctx.Abort()
		return
	}
	ctx.Next()
}

// insecureCookies is solely for local HTTP development. Production must use HTTPS.
func insecureCookies() bool { return os.Getenv("KEIJI_INSECURE_COOKIE") == "true" }

// SameOrigin protects unsafe browser requests, including ordinary forms and HTMX.
// Non-browser clients must supply an Origin matching the public request URL.
func SameOrigin(ctx *gin.Context) {
	switch ctx.Request.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		ctx.Next()
		return
	}
	source := ctx.GetHeader("Origin")
	origin := source != ""
	if !origin {
		source = ctx.GetHeader("Referer")
	}
	u, err := url.Parse(source)
	scheme := "https"
	if insecureCookies() {
		scheme = "http"
	}
	if err != nil || u == nil || u.User != nil || u.Opaque != "" ||
		!strings.EqualFold(u.Scheme, scheme) || !strings.EqualFold(u.Host, ctx.Request.Host) ||
		u.Host == "" || (origin && (u.Path != "" || u.RawQuery != "" || u.Fragment != "")) {
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}
	ctx.Next()
}

type loginWindow struct {
	count int
	until time.Time
}

// LoginRateLimit limits all attempts, not just failures, so concurrent attempts
// cannot race a post-authentication counter. State is local to this process.
func LoginRateLimit() gin.HandlerFunc {
	var mu sync.Mutex
	windows := make(map[string]loginWindow)
	var proxies []*net.IPNet
	for _, value := range strings.Split(os.Getenv("KEIJI_TRUSTED_PROXIES"), ",") {
		value = strings.TrimSpace(value)
		if ip := net.ParseIP(value); ip != nil {
			bits := 128
			if ip.To4() != nil {
				ip = ip.To4()
				bits = 32
			}
			proxies = append(proxies, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
		} else if _, network, err := net.ParseCIDR(value); err == nil {
			proxies = append(proxies, network)
		}
	}
	trusted := func(ip net.IP) bool {
		for _, network := range proxies {
			if network.Contains(ip) {
				return true
			}
		}
		return false
	}
	return func(ctx *gin.Context) {
		peer, _, err := net.SplitHostPort(ctx.Request.RemoteAddr)
		if err != nil {
			peer = ctx.Request.RemoteAddr
		}
		ip := net.ParseIP(peer)
		if ip != nil && trusted(ip) {
			// Walk from the trusted peer toward the client, stopping at the first
			// untrusted hop. Never accept forwarding headers from an untrusted peer.
			parts := strings.Split(ctx.GetHeader("X-Forwarded-For"), ",")
			for i := len(parts) - 1; i >= 0 && trusted(ip); i-- {
				next := net.ParseIP(strings.TrimSpace(parts[i]))
				if next == nil {
					break
				}
				ip = next
			}
		}
		key := "unknown"
		if ip != nil {
			key = ip.String()
		}
		now := time.Now()
		mu.Lock()
		w := windows[key]
		if !now.Before(w.until) {
			w = loginWindow{until: now.Add(10 * time.Minute)}
		}
		if w.count >= 10 {
			mu.Unlock()
			ctx.Header("Retry-After", "600")
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		if _, exists := windows[key]; !exists && len(windows) >= 10000 {
			for k, old := range windows {
				if !now.Before(old.until) {
					delete(windows, k)
				}
			}
			if len(windows) >= 10000 {
				mu.Unlock()
				ctx.Header("Retry-After", "600")
				ctx.AbortWithStatus(http.StatusTooManyRequests)
				return
			}
		}
		w.count++
		windows[key] = w
		mu.Unlock()
		ctx.Next()
	}
}
