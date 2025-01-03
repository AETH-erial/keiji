package auth

import (
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
)

type InvalidCredentials struct{}

func (i *InvalidCredentials) Error() string {
	return "Invalid credentials supplied."
}

type Credentials struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

type AuthCache struct {
	AuthCookies *cache.Cache
}

const (
	defaultExpiration = 20 * time.Minute
	purgeTime         = 1 * time.Hour
)

func NewCache() *AuthCache {
	Cache := cache.New(defaultExpiration, purgeTime)
	return &AuthCache{
		AuthCookies: Cache,
	}
}

func (c *AuthCache) update(id string, cookie string) {
	c.AuthCookies.Set(id, cookie, cache.DefaultExpiration)
}

func (c *AuthCache) Read(id string) bool {
	_, ok := c.AuthCookies.Get(id)
	return ok
}

/*
Recieve the credentials from frontend and validate them

	:param c: pointer to Credential struct
*/
func Authorize(c *Credentials, cache *AuthCache) (string, error) {
	if c.Username == "" || c.Password == "" {
		return "", &InvalidCredentials{}
	}
	if c.Username == os.Getenv("USERNAME") {
		if c.Password == os.Getenv("PASSWORD") {
			id := uuid.New()
			cache.update(id.String(), id.String())
			return id.String(), nil
		}
		return "", &InvalidCredentials{}

	}
	return "", &InvalidCredentials{}
}
