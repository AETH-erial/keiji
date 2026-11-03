# Offline legacy image migration

Stop Keiji **and every other database and media writer** before starting. Take and verify a restorable backup of the SQLite database (including any WAL/SHM state; use SQLite's backup facilities or a cleanly closed database) **and** media directory. Store it elsewhere. This tool does not create or verify backups and does not establish that the application is offline. Restore the backup if recovery is needed.

```sh
go run ./cmd/keiji-migrate -db /path/to/sqlite.db -media-dir /path/to/images \
  -legacy-category digital_art -offline -backup-confirmed
```

The acknowledgements are mandatory. Choose the category for all legacy rows only when `images.category` is absent; existing categories are never overwritten. Missing `category` and `media_type` columns are added and backfill runs in one SQLite transaction. SQLite must be writable, the database must already exist, and `images` must have an `id` and a rowid. The tool does not modify media files. Repeated runs preserve populated media types and retry unknowns, allowing missing files to be restored later.

Detection uses content, not extensions: JPEG, PNG, GIF, WebP and WebM (an EBML header with WebM DocType). Only canonical lowercase UUID filenames are opened; symlinks and non-regular files are ignored. At most 1 MiB is read per file; exceptionally large EBML headers are unknown. Missing, invalid-name, symlink, and unrecognized files are counted as unknown; unexpected filesystem or database errors roll back the transaction. Newly added media type columns store unknown as NULL. If an existing column is NOT NULL with empty-string unknowns, those values remain empty instead of violating its constraint. Existing populated values are never re-detected. Output counts only rows with missing media types, not already-populated rows.

Run tests with `go test ./pkg/storage ./cmd/keiji-migrate`.
