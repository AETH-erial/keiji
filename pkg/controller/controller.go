package controller

import (
	"io/fs"

	"git.aetherial.dev/aeth/keiji/pkg/helpers"
)

type Controller struct {
	Domain      string
	database    helpers.DocumentIO
	RedisConfig helpers.RedisConf
	Cache       *helpers.AllCache
	FileIO      fs.FS
}

func NewController(domain string, redisPort string, redisAddr string, database helpers.DocumentIO, files fs.FS) *Controller {
	return &Controller{Cache: helpers.NewCache(),
		Domain: domain, RedisConfig: helpers.RedisConf{
			Port: redisPort,
			Addr: redisAddr,
		},
		database: database,
		FileIO:   files,
	}
}
