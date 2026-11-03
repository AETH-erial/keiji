package storage

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

type legacyImageIO struct {
	reads int
	data  []byte
	err   error
}

func (l *legacyImageIO) Put(b []byte, id Identifier) error { l.data = b; return nil }
func (l *legacyImageIO) Get(id Identifier) ([]byte, error) { l.reads++; return l.data, l.err }

func TestImageListsMetadataAndLegacyOpen(t *testing.T) {
	repo, db := newTestDb(t.TempDir(), true)
	defer db.Close()
	legacy := &legacyImageIO{data: []byte("0123456789")}
	repo.imageIO = legacy
	_, err := db.Exec(`INSERT INTO images (id,title,desc,created,category,media_type) VALUES ('x','title','desc','today','digital_art','video/webm')`)
	require.NoError(t, err)
	all := repo.GetAllImages()
	category := repo.GetImagesByCategory(DIGITAL_ART)
	require.Len(t, all, 1)
	require.Equal(t, all, category)
	require.Nil(t, all[0].Data)
	require.Equal(t, 0, legacy.reads)
	img, stream, err := repo.OpenImage("x")
	require.NoError(t, err)
	defer stream.Close()
	require.Equal(t, "video/webm", img.MediaType)
	require.Nil(t, img.Data)
	_, err = stream.Seek(2, io.SeekStart)
	require.NoError(t, err)
	b, err := io.ReadAll(stream)
	require.NoError(t, err)
	require.Equal(t, "23456789", string(b))
	require.Equal(t, 1, legacy.reads)
	legacy.err = errors.New("failed")
	_, _, err = repo.OpenImage("x")
	require.ErrorIs(t, err, legacy.err)
	_, _, err = repo.OpenImage("missing")
	require.ErrorIs(t, err, ErrNotExists)
	_, _, err = repo.OpenImage("../escape")
	require.ErrorIs(t, err, ErrNotExists)
}

func TestFilesystemImageOpen(t *testing.T) {
	f := FilesystemImageIO{RootDir: t.TempDir()}
	require.NoError(t, f.Put([]byte("0123456789"), "x"))
	require.NoError(t, f.Put([]byte("abc"), "x"))
	stream, err := f.Open("x")
	require.NoError(t, err)
	defer stream.Close()
	b, err := io.ReadAll(stream)
	require.NoError(t, err)
	require.Equal(t, "abc", string(b))
	_, err = f.Open("missing")
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = f.Open("../escape")
	require.Error(t, err)
}

func TestImageListQueryFailure(t *testing.T) {
	repo, db := newTestDb(t.TempDir(), true)
	require.NoError(t, db.Close())
	require.Nil(t, repo.GetAllImages())
	require.Nil(t, repo.GetImagesByCategory(DIGITAL_ART))
}
