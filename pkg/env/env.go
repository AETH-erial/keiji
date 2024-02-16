package env

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

const IMAGE_STORE = "IMAGE_STORE"
const HOST_PORT = "HOST_PORT"
const HOST_ADDR = "HOST_ADDR"

var REQUIRED_VARS = []string{
	IMAGE_STORE,
	HOST_PORT,
	HOST_ADDR,

}

type EnvNotSet struct {
	NotSet string
}

func (e *EnvNotSet) Error() string {
	return fmt.Sprintf("Environment variable: '%s' was not set.", e.NotSet)
}

/*
  verify all environment vars passed in are set
          :param vars: array of strings to verify
*/
func LoadAndVerifyEnv(path string, vars []string) error {

	err := godotenv.Load(path)
	if err != nil {
		return err
	}

	for i := range vars {
		if os.Getenv(vars[i]) == "" {
			return &EnvNotSet{NotSet: vars[i]}
		}
	}
	return nil
}
