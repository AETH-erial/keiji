package main

import (
	"log"
	"os"

	"git.aetherial.dev/aeth/keiji/pkg/env"
	"git.aetherial.dev/aeth/keiji/pkg/helpers"
)


func main() {
	err := env.LoadAndVerifyEnv(os.Args[1], env.REQUIRED_VARS); if err != nil {
		log.Fatal(err)
	}
	rds := helpers.NewRedisClient(helpers.RedisConf{Port: os.Getenv("REDIS_PORT"), Addr: os.Getenv("REDIS_ADDR")})
	err = rds.SeedData(os.Args[2]); if err != nil {
		log.Fatal(err)
	}
	
}