package env

import (
	"io"
	"os"
	"path"
	"strings"
	"testing"
)

// Testing that the template writer creates the appropriate data
func TestWriteTemplate(t *testing.T) {
	tmp := t.TempDir()
	envTmp := path.Join(tmp, ".env")
	wr, err := os.Create(envTmp)
	if err != nil {
		t.Errorf("failed to created test file: %s\n", err.Error())
	}
	defer wr.Close()
	err = WriteTemplate(wr)
	if err != nil {
		t.Errorf("failed to write to test file: %s\n", err.Error())
	}
	wr.Close()
	r, err := os.Open(envTmp)
	if err != nil {
		t.Errorf("failed to open test file for reading: %s\n", err.Error())
	}
	got, err := io.ReadAll(r)
	if err != nil {
		t.Errorf("failed to read from open file: %s\n", err.Error())
	}

	want := []string{
		KEIJI_USERNAME,
		KEIJI_PASSWORD,
		IMAGE_STORE,
		WEB_ROOT,
		DOMAIN_NAME,
		HOST_PORT,
		HOST_ADDR,
		USE_SSL,
		CHAIN,
		KEY,
	}
	for i := range want {

		if !strings.Contains(string(got), want[i]) {
			t.Errorf("test failed! Template file with content: '%s'\n Did not contain: '%s'\n", string(got), want[i])
		}
	}

}
