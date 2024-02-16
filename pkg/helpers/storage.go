package helpers

import (
	"os"

	"git.aetherial.dev/aeth/keiji/pkg/env"
)




func GetImageStore() string {
	return os.Getenv(env.IMAGE_STORE)
}

