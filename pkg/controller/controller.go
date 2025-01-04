package controller

import (
	"io/fs"

	"git.aetherial.dev/aeth/keiji/pkg/auth"
	"git.aetherial.dev/aeth/keiji/pkg/storage"
)

type Controller struct {
	Domain     string
	database   storage.DocumentIO
	Cache      *auth.AuthCache
	AuthSource auth.Source
	FileIO     fs.FS
}

func NewController(domain string, database storage.DocumentIO, files fs.FS, authSrc auth.Source) *Controller {
	return &Controller{
		Cache:      auth.NewCache(),
		AuthSource: authSrc,
		Domain:     domain,
		database:   database,
		FileIO:     files,
	}
}
