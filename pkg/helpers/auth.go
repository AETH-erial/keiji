package helpers

import (
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
)


type InvalidCredentials struct {}

func (i *InvalidCredentials) Error() string {
	return "Invalid credentials supplied."
}


type Credentials struct {
	Username	string	`form:"username" json:"username"`
	Password	string	`form:"password" json:"password"`
}

type AllCache struct {
    AuthCookies *cache.Cache
}

const (
    defaultExpiration = 20 * time.Minute
    purgeTime         = 1 * time.Hour
)

func NewCache() *AllCache {
    Cache := cache.New(defaultExpiration, purgeTime)
    return &AllCache{
        AuthCookies: Cache,
    }
}

func (c *AllCache) update(id string, cookie string) {
    c.AuthCookies.Set(id, cookie, cache.DefaultExpiration)
}

func (c *AllCache) Read(id string) bool {
    _, ok := c.AuthCookies.Get(id)
    if ok {

        return true
    }
    return false
}


/*
Recieve the credentials from frontend and validate them
	:param c: pointer to Credential struct
*/
func Authorize(c *Credentials, cache *AllCache) (string, error) {
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
