package controller

import (
	"git.aetherial.dev/aeth/keiji/pkg/helpers"
)

type Controller struct {
	Domain      string
	database    helpers.DocumentIO
	RedisConfig helpers.RedisConf
	Cache       *helpers.AllCache
}

func NewController(domain string, redisPort string, redisAddr string, database helpers.DocumentIO) *Controller {
	return &Controller{Cache: helpers.NewCache(),
		Domain: domain, RedisConfig: helpers.RedisConf{
			Port: redisPort,
			Addr: redisAddr,
		},
		database: database,
	}
}
