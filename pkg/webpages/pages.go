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
)

//go:embed html
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
		return content
	}
	if opt == FILESYSTEM {
		fmt.Println(os.Getenv("WEB_ROOT"))
		return FilesystemWebpages{Webroot: path.Base(os.Getenv("WEB_ROOT"))}
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
	fmt.Println(path.Join(f.Webroot, file))
	return os.Open(path.Join(f.Webroot, file))
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
