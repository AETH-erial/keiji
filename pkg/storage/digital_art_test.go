package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDigitalArtMetadataUpdate(t *testing.T) {
	dir := t.TempDir()
	repo, db := newTestDb(dir, true)
	defer db.Close()
	id, err := repo.AddImage([]byte("image bytes"), "old", "old desc", DIGITAL_ART, "image/png")
	require.NoError(t, err)
	other, err := repo.AddImage([]byte("other"), "other", "desc", MISC_ASSET, "image/jpeg")
	require.NoError(t, err)
	before, err := repo.GetImage(id)
	require.NoError(t, err)
	require.NoError(t, repo.UpdateDigitalArt(id, "new", "new desc"))
	after, err := repo.GetImage(id)
	require.NoError(t, err)
	require.Equal(t, "new", after.Title)
	require.Equal(t, "new desc", after.Desc)
	require.Equal(t, before.Data, after.Data)
	require.Equal(t, before.Created, after.Created)
	require.Equal(t, before.MediaType, after.MediaType)
	require.Equal(t, before.Category, after.Category)
	require.ErrorIs(t, repo.UpdateDigitalArt(other, "bad", "bad"), ErrNotExists)
	otherAfter, err := repo.GetImage(other)
	require.NoError(t, err)
	require.Equal(t, "other", otherAfter.Title)
	require.Equal(t, "desc", otherAfter.Desc)
	require.ErrorIs(t, repo.UpdateDigitalArt("missing", "x", "y"), ErrNotExists)
}

func TestDigitalArtUpdateSafety(t *testing.T) {
	dir := t.TempDir()
	repo, db := newTestDb(dir, true)
	defer db.Close()
	id, err := repo.AddImage([]byte("preserve"), "art", "desc", DIGITAL_ART, "image/png")
	require.NoError(t, err)
	// A duplicate ID can share the same file, even across categories.
	_, err = db.Exec("INSERT INTO images (id,title,desc,created,category) VALUES (?, 'other', '', '', ?)", id, MISC_ASSET)
	require.NoError(t, err)
	require.ErrorIs(t, repo.UpdateDigitalArt(id, "bad", "bad"), ErrNotExists)
	data, err := os.ReadFile(filepath.Join(dir, string(id)))
	require.NoError(t, err)
	require.Equal(t, []byte("preserve"), data)
	var count int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM images WHERE id = ?", id).Scan(&count))
	require.Equal(t, 2, count)
	for _, bad := range []Identifier{"", ".", "..", "../outside", "nested/name", `..\outside`} {
		require.Error(t, repo.UpdateDigitalArt(bad, "x", "y"))
	}
}

func TestArchiveDigitalArtPreservesMedia(t *testing.T) {
	dir := t.TempDir()
	repo, db := newTestDb(dir, true)
	defer db.Close()
	id, err := repo.AddImage([]byte("image bytes"), "art", "description", DIGITAL_ART, "image/png")
	require.NoError(t, err)
	before, err := repo.GetImage(id)
	require.NoError(t, err)

	require.NoError(t, repo.ArchiveDigitalArt(id))
	after, err := repo.GetImage(id)
	require.NoError(t, err)
	before.Category = MISC_ASSET
	require.Equal(t, before, after)
	data, err := os.ReadFile(filepath.Join(dir, string(id)))
	require.NoError(t, err)
	require.Equal(t, []byte("image bytes"), data)
	art, err := repo.GetImagesByCategoryWithError(DIGITAL_ART)
	require.NoError(t, err)
	require.Empty(t, art)
	misc, err := repo.GetImagesByCategoryWithError(MISC_ASSET)
	require.NoError(t, err)
	require.Len(t, misc, 1)
	require.Equal(t, id, misc[0].Ident)
	require.ErrorIs(t, repo.ArchiveDigitalArt(id), ErrNotExists)
	require.ErrorIs(t, repo.ArchiveDigitalArt("missing"), ErrNotExists)
}

func TestArchiveDigitalArtSafety(t *testing.T) {
	dir := t.TempDir()
	repo, db := newTestDb(dir, true)
	defer db.Close()
	other, err := repo.AddImage([]byte("other"), "other", "desc", MISC_ASSET, "image/jpeg")
	require.NoError(t, err)
	require.ErrorIs(t, repo.ArchiveDigitalArt(other), ErrNotExists)
	id, err := repo.AddImage([]byte("preserve"), "art", "desc", DIGITAL_ART, "image/png")
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO images (id,title,desc,created,category) VALUES (?, 'other', '', '', ?)", id, MISC_ASSET)
	require.NoError(t, err)
	require.ErrorIs(t, repo.ArchiveDigitalArt(id), ErrNotExists)
	var artCount, total int
	require.NoError(t, db.QueryRow("SELECT COUNT(*), COUNT(CASE WHEN category = ? THEN 1 END) FROM images WHERE id = ?", DIGITAL_ART, id).Scan(&total, &artCount))
	require.Equal(t, 2, total)
	require.Equal(t, 1, artCount)
	data, err := os.ReadFile(filepath.Join(dir, string(id)))
	require.NoError(t, err)
	require.Equal(t, []byte("preserve"), data)
	for _, bad := range []Identifier{"", ".", "..", "../outside", "nested/name", `..\outside`} {
		require.Error(t, repo.ArchiveDigitalArt(bad))
	}
}

func TestArchiveDigitalArtDatabaseFailure(t *testing.T) {
	repo, db := newTestDb(t.TempDir(), true)
	require.NoError(t, db.Close())
	require.Error(t, repo.ArchiveDigitalArt("valid-id"))
}
