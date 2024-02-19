package helpers

import (
	"fmt"
	"os"

	"git.aetherial.dev/aeth/keiji/pkg/env"
)

type InvalidSkipArg struct {Skip int}

func (i *InvalidSkipArg) Error() string {
	return fmt.Sprintf("Invalid skip amount was passed: %s", i.Skip)
}

/*
Function to return the location of the image store. Wrapping the env call in
a function so that refactoring is easier
*/
func GetImageStore() string {
	return os.Getenv(env.IMAGE_STORE)
}

/*
Return all of the filenames of the images that exist in the imagestore location
	:param limit: the limit of filenames to return
	:param skip: the index to start getting images from
*/
func GetImagePaths(limit int, skip int) ([]string, error) {
	f, err := os.ReadDir(GetImageStore())
	if err != nil {
		return nil, err
	}
	if len(f) < skip {
		return nil, &InvalidSkipArg{Skip: skip}
	}
	if len(f) < limit {
		return nil, &InvalidSkipArg{Skip: limit}
	}
	fnames := []string{}
	for i := skip; i < (skip + limit); i++ {
		fnames = append(fnames, fmt.Sprintf("/api/v1/images/%s", f[i].Name()))
	}
	return fnames, err
}

