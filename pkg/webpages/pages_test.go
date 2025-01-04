package webpages

import (
	"io"
	"io/fs"
	"os"
	"path"
	"testing"

	"git.aetherial.dev/aeth/keiji/pkg/env"
	"github.com/stretchr/testify/assert"
)

func TestNewContentLayer(t *testing.T) {
	type testcase struct {
		input ServiceOption
		want  fs.FS
	}
	for _, tc := range []testcase{
		{
			input: EMBED,
			want:  content,
		},
		{
			input: FILESYSTEM,
			want:  FilesystemWebpages{Webroot: path.Base(os.Getenv(env.WEB_ROOT))},
		},
	} {
		got := NewContentLayer(tc.input)
		assert.Equal(t, tc.want, got)

	}
}

func TestOpen(t *testing.T) {

	type testcase struct {
		input      string
		dir        string
		createFile string
		data       []byte
		err        error
	}
	for _, tc := range []testcase{
		{
			input:      "testfile.txt",
			dir:        t.TempDir(),
			createFile: "testfile.txt",
			data:       []byte("testdataetcabc123"),
			err:        nil,
		},
		{
			input:      "this_cant_be_indexed.csv",
			dir:        t.TempDir(),
			createFile: "test",
			data:       []byte("testdatagdfsagfdsbhs"),
			err:        os.ErrNotExist,
		},
	} {
		os.WriteFile(path.Join(tc.dir, tc.createFile), tc.data, os.ModePerm)
		testFs := FilesystemWebpages{Webroot: tc.dir}
		fh, err := testFs.Open(tc.input)
		if err != nil {
			assert.Equal(t, tc.err, err)
		} else {
			b, _ := io.ReadAll(fh)
			assert.Equal(t, tc.data, b)
		}

	}

}

func TestReadToString(t *testing.T) {

	testFs := FilesystemWebpages{Webroot: t.TempDir()}
	filename := "testfile.txt"
	testFile := path.Join(testFs.Webroot, filename)
	data := []byte("abc123xyz098")
	os.WriteFile(testFile, data, os.ModePerm)
	got := ReadToString(testFs, filename)
	assert.Equal(t, string(data), got)

}
