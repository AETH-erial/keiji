package env

import (
	"bytes"
	"io"
	"log"
	"os"
	"testing"
)

// Testing that the template writer creates the appropriate data
func TestWriteTemplate(t *testing.T) {
	b, err := os.ReadFile("../../test/.env.template")
	if err != nil {
		log.Fatal(err)
	}
	buf := bytes.NewBuffer([]byte{})
	err = WriteTemplate(buf)
	if err != nil {
		log.Fatal(err)
	}
	got, err := io.ReadAll(buf)
	if string(got) != string(b) {
		t.Errorf("test failed! Got: %s\nWanted: %s\n", string(got), string(b))
	}

}
