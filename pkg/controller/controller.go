package controller

import (
	"io/fs"

	"git.aetherial.dev/aeth/keiji/pkg/auth"
	"git.aetherial.dev/aeth/keiji/pkg/storage"
)

type Controller struct {
	Domain   string
	database storage.DocumentIO
	Cache    *auth.AuthCache
	FileIO   fs.FS
}

func NewController(domain string, database storage.DocumentIO, files fs.FS) *Controller {
	return &Controller{
		Cache:    auth.NewCache(),
		Domain:   domain,
		database: database,
		FileIO:   files,
	}
}
