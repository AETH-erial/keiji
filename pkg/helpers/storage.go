package helpers

import (
	"os"

	"adeptus-mechanicus.void/git/keiji/pkg/env"
)




func GetImageStore() string {
	return os.Getenv(env.IMAGE_STORE)
}

