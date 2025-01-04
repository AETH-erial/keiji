package webpages

import (
	"embed"
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"

	"git.aetherial.dev/aeth/keiji/pkg/env"
)

//go:embed html cdn
var content embed.FS

type ServiceOption string

const EMBED ServiceOption = "embed"
const FILESYSTEM ServiceOption = "filesystem"

/*
Creates the new filesystem implementer for serving the webpages to the API

	:param opt: the service option to
*/
func NewContentLayer(opt ServiceOption) fs.FS {
	if opt == EMBED {
		fmt.Println("Using embed files to pull html templates")
		return content
	}
	if opt == FILESYSTEM {
		fmt.Println("Using filesystem to pull html templates")

		return FilesystemWebpages{Webroot: path.Base(os.Getenv(env.WEB_ROOT))}
	}
	log.Fatal("Unknown option was passed: ", opt)
	return content

}

type WebContentLayer interface{}

type EmbeddedWebpages struct{}

type FilesystemWebpages struct {
	Webroot string
}

/*
Implementing the io.FS interface for interoperability
*/
func (f FilesystemWebpages) Open(file string) (fs.File, error) {
	filePath := path.Join(f.Webroot, file)
	fh, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening the file: %s because %s", filePath, err)
		return nil, os.ErrNotExist
	}
	return fh, nil
}

/*
Read content to a string for easy template ingestion. Will panic if the underlying os.Open call fails
*/
func ReadToString(rdr fs.FS, name string) string {
	fh, err := rdr.Open(name)
	if err != nil {
		log.Fatal(err, "couldnt open the file: ", name)
	}
	b, err := io.ReadAll(fh)

	if err != nil {
		log.Fatal("Could not read the file: ", name)
	}
	return string(b)

}
