package controller

import (
	"io"
	"io/fs"
	"os"

	"git.aetherial.dev/aeth/keiji/pkg/auth"
	"git.aetherial.dev/aeth/keiji/pkg/storage"
)

type Controller struct {
	Domain     string
	database   storage.DocumentIO
	Cache      *auth.AuthCache
	AuthSource auth.Source
	FileIO     fs.FS
	log        io.Writer
}

func NewController(domain string, database storage.DocumentIO, files fs.FS, authSrc auth.Source) *Controller {
	return &Controller{
		Cache:      auth.NewCache(),
		AuthSource: authSrc,
		Domain:     domain,
		database:   database,
		FileIO:     files,
		log:        os.Stdout,
	}
}

func (c Controller) LogMsg(msg ...string) {
	for i := range msg {
		c.log.Write([]byte(msg[i]))
	}
	c.log.Write([]byte("\n"))
}
