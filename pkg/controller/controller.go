package controller

import (
	"io/fs"

	"git.aetherial.dev/aeth/keiji/pkg/helpers"
)

type Controller struct {
	Domain   string
	database helpers.DocumentIO
	Cache    *helpers.AuthCache
	FileIO   fs.FS
}

func NewController(domain string, database helpers.DocumentIO, files fs.FS) *Controller {
	return &Controller{Cache: helpers.NewCache(),
		Domain:   domain,
		database: database,
		FileIO:   files,
	}
}
