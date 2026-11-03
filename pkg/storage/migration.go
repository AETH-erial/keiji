package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/google/uuid"
)

// MigrationReport counts rows examined and their outcomes. Unknown includes missing,
// unsafe, unreadable-by-format, or unsupported media, but not I/O errors.
type MigrationReport struct{ Examined, Updated, Unknown int }

// MigrateLegacyImages adds legacy image columns and backfills missing media types.
// The caller must stop all database and media writers and create a verified backup
// first. All database changes are committed together. Existing values are preserved.
// category is the intended category for rows when that column is absent.
func MigrateLegacyImages(ctx context.Context, db *sql.DB, mediaDir, category string) (MigrationReport, error) {
	var report MigrationReport
	if strings.TrimSpace(category) == "" {
		return report, errors.New("legacy category must not be empty")
	}
	info, err := os.Stat(mediaDir)
	if err != nil {
		return report, fmt.Errorf("media directory: %w", err)
	}
	if !info.IsDir() {
		return report, errors.New("media path is not a directory")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return report, err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `PRAGMA table_info(images)`)
	if err != nil {
		return report, err
	}
	columns := map[string]bool{}
	for rows.Next() {
		var cid, notnull, pk int
		var name, typ string
		var def sql.NullString
		if err = rows.Scan(&cid, &name, &typ, &notnull, &def, &pk); err != nil {
			rows.Close()
			return report, err
		}
		columns[strings.ToLower(name)] = true
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return report, err
	}
	if !columns["id"] {
		return report, errors.New("images table missing or has no id column")
	}
	if !columns["category"] {
		if _, err = tx.ExecContext(ctx, `ALTER TABLE images ADD COLUMN category TEXT`); err != nil {
			return report, err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE images SET category = ?`, category); err != nil {
			return report, err
		}
	}
	if !columns["media_type"] {
		if _, err = tx.ExecContext(ctx, `ALTER TABLE images ADD COLUMN media_type TEXT`); err != nil {
			return report, err
		}
	}
	rows, err = tx.QueryContext(ctx, `SELECT rowid, id FROM images WHERE media_type IS NULL OR media_type = ''`)
	if err != nil {
		return report, err
	}
	type item struct {
		row int64
		id  sql.NullString
	}
	var items []item
	for rows.Next() {
		var i item
		if err = rows.Scan(&i.row, &i.id); err != nil {
			rows.Close()
			return report, err
		}
		items = append(items, i)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return report, err
	}
	for _, i := range items {
		report.Examined++
		mediaType, e := migrationMediaType(mediaDir, i.id.String)
		if e != nil {
			return MigrationReport{}, fmt.Errorf("media for row %d: %w", i.row, e)
		}
		if mediaType == "" {
			report.Unknown++
			continue
		}
		if _, err = tx.ExecContext(ctx, `UPDATE images SET media_type = ? WHERE rowid = ?`, mediaType, i.row); err != nil {
			return MigrationReport{}, err
		}
		report.Updated++
	}
	if err = tx.Commit(); err != nil {
		return MigrationReport{}, err
	}
	return report, nil
}

// Read only canonical UUID filenames, never follow symlinks, and bound memory.
func migrationMediaType(dir, name string) (string, error) {
	id, err := uuid.Parse(name)
	if err != nil || id.String() != name {
		return "", nil
	}
	fd, err := syscall.Open(filepath.Join(dir, name), syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ELOOP) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	f := os.NewFile(uintptr(fd), name)
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", nil
	}
	data, err := io.ReadAll(io.LimitReader(f, 1<<20))
	if err != nil {
		return "", err
	}
	return detectMediaType(data), nil
}
