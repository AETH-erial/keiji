package controller

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"git.aetherial.dev/aeth/keiji/pkg/storage"
	"github.com/gin-gonic/gin"
)

type trackedImageStream struct {
	*bytes.Reader
	closed bool
}

func (s *trackedImageStream) Close() error { s.closed = true; return nil }

type streamingRepo struct {
	*mediaRepo
	stream *trackedImageStream
	err    error
}

func (r *streamingRepo) OpenImage(id storage.Identifier) (storage.Image, storage.ReadSeekCloser, error) {
	return storage.Image{MediaType: "video/webm"}, r.stream, r.err
}
func TestServeImageStream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		method, rangeHeader string
		status              int
		body                string
	}{
		{"GET", "bytes=2-5", 206, "2345"}, {"HEAD", "", 200, ""}, {"GET", "bytes=20-30", 416, ""},
	} {
		stream := &trackedImageStream{Reader: bytes.NewReader([]byte("0123456789"))}
		repo := &streamingRepo{mediaRepo: &mediaRepo{}, stream: stream}
		router := gin.New()
		c := &Controller{database: repo}
		router.Handle(tc.method, "/images/:file", c.ServeImage)
		req := httptest.NewRequest(tc.method, "/images/x", nil)
		req.Header.Set("Range", tc.rangeHeader)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != tc.status || (tc.status != 416 && res.Body.String() != tc.body) || !stream.closed {
			t.Fatalf("%s %d %q closed=%v", tc.method, res.Code, res.Body.String(), stream.closed)
		}
	}
}

type trackingFS struct{ file *trackingFile }

func (f trackingFS) Open(name string) (fs.File, error) { return f.file, nil }

type trackingFile struct {
	fs.File
	closed bool
	fail   bool
}

func (f *trackingFile) Close() error { f.closed = true; return f.File.Close() }
func (f *trackingFile) Read(b []byte) (int, error) {
	if f.fail {
		return 0, errors.New("read failed")
	}
	return f.File.Read(b)
}
func TestServeGenericClosesFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, fail := range []bool{false, true} {
		f, err := (fstest.MapFS{"cdn/a.js": &fstest.MapFile{Data: []byte("test")}}).Open("cdn/a.js")
		if err != nil {
			t.Fatal(err)
		}
		tf := &trackingFile{File: f, fail: fail}
		c := &Controller{FileIO: trackingFS{tf}}
		router := gin.New()
		router.GET("/cdn/:file", c.ServeGeneric)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/cdn/a.js", nil))
		want := 200
		if fail {
			want = 500
		}
		if res.Code != want || !tf.closed {
			t.Fatalf("status %d closed %v", res.Code, tf.closed)
		}
	}
}

var _ io.ReadSeeker = (*trackedImageStream)(nil)
