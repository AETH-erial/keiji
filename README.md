## *keiji*, SSR blog + website using just HTMX and Golang

This is my personal blogging site, in lieu of having social media like Facebook, Instagram, Tumblr, Twitter, etc. I have handrolled my own blogging platform to be used exclusively by me. Designed to be my 'virtual studio', where I can put my art, writings, and various updates onto the internet whilst retaining 99% control over the platform wherein they reside.

### Highlights
- clean contract between the webserver and the database
- embedded web content to make deployments easier
- dynamic population of navigation bars, menu, and the admin panel
- cool mid 2000s website feel


### Deployment and persistent storage

Set the required environment variables `HOST_ADDR`, `HOST_PORT`, `DOMAIN_NAME`,
`USE_SSL`, `KEIJI_USERNAME`, and `KEIJI_PASSWORD`. `IMAGE_STORE` must also be
set to a writable directory for uploaded images. The `-env` flag selects an
environment file (default `.env`); existing process environment variables take
precedence. Run `keiji -blank` to generate an environment template, then start
with `keiji -content embed` (or `-content fs` with `WEB_ROOT` configured).
Protect the environment file and credentials from unauthorized access.

`SQLITE_PATH` optionally selects the SQLite database file; when unset or empty,
it defaults to `sqlite.db` in the process working directory, as in older
versions. Set it to an absolute path such as `/data/sqlite.db` for deployment.
The parent directory must already exist and be writable by the service; SQLite
may create journal, WAL, and shared-memory files alongside the database. Moving
an existing installation requires moving the existing database to the configured
path, not starting with an empty database. Protect database and image files
with appropriate filesystem permissions.

For containers, mount persistent storage for **both** the database directory
and `IMAGE_STORE`, for example mount a volume at `/data`, set
`SQLITE_PATH=/data/sqlite.db` and `IMAGE_STORE=/data/images`, and ensure the
container user can write there. Do not rely on the container writable layer or
working directory surviving replacement. Keep a single application instance
writing to these files; SQLite locking and image storage require a filesystem
with suitable local locking semantics.

### Backup and restore

Back up both the SQLite database and the entire image store: image metadata and
image files are separate, and either alone is incomplete. The simplest consistent
procedure is to stop all Keiji writers, copy the database and image directory
as one backup, then restart Keiji. If using an online SQLite backup mechanism,
coordinate it with image writes or quiesce the application to obtain a matching
snapshot; copying a live database file alone is not a reliable SQLite backup,
especially in WAL mode. Keep backups off-host, secure them, and periodically
test restoration. To restore, stop Keiji, restore the matching database and
image directory to the configured paths, ensure ownership and permissions, then
restart and verify content. Preserve or correctly checkpoint any SQLite
sidecar files before moving a database.

### HTTPS

For direct TLS set `USE_SSL=true`, `CHAIN` to the certificate chain, and `KEY`
to its private key. Listener errors, including invalid certificates, abort startup
without falling back to HTTP. Alternatively terminate HTTPS at a trusted reverse
proxy and set `USE_SSL=false` for a protected backend connection. See
[pkg/controller/SECURITY.md](pkg/controller/SECURITY.md) for cookie, proxy,
CSRF and rate-limiting deployment requirements; do not use insecure cookies in
production.

### Migrating older databases

For legacy image tables or missing media types, use the offline, repeatable
[`keiji-migrate` command](cmd/keiji-migrate/README.md). Stop all writers and
back up both the database and media directory first. Choose the legacy category
based on the old installation's actual content; the tool does not guess it.

### Service account setup

`sudo ./deploy/setup.sh` creates the `keiji` system group and a system user with
no home directory created and `/sbin/nologin` as its shell. It creates
`/etc/keiji` as `root:keiji` mode `0750` and `/var/lib/keiji` plus its `images`
subdirectory as `keiji:root` mode `0700`. Root manages configuration; the service
can read it but cannot change it. Install the actual environment file separately:

```sh
sudo install -o root -g keiji -m 0640 /path/to/keiji.env /etc/keiji/keiji.env
```

Only put the service account and trusted administrators in the `keiji` group.
The script does not recursively change existing files, ACLs, or mounted
filesystems: inspect and fix any pre-existing contents and ACLs before use.
It does not install binaries, create a database, or enable/start a service. An
existing `keiji` account with a different primary group or shell must be reviewed
manually; the script will not silently repurpose it.

### Install with systemd

Build and test as an ordinary user, then install as root:

```sh
make build
make test
sudo make install
```

The install target runs `deploy/setup.sh`, installs `keiji` and `keiji-ctl` in
`/usr/local/bin`, and replaces `/etc/systemd/system/keiji.service` with the
repo's unit. It does not restart or enable the service, and does not create or
replace the env file, database, or media. Put local unit customizations in a
systemd drop-in (`systemctl edit keiji`) so reinstalls won't overwrite them.

For a first installation, edit a private copy of `deploy/keiji.env.example` and
install it as `/etc/keiji/keiji.env` with the permissions above. Replace the
example credentials. The example binds loopback for TLS termination by Caddy;
see the security configuration for reverse-proxy requirements. The service's
`ProtectSystem=strict` allows writes only under `/var/lib/keiji`, so other data
locations need an appropriate systemd drop-in. The unit does not load the file
via systemd `EnvironmentFile`; Keiji reads it directly via `-env`.

After installing or changing the unit, run `sudo systemctl daemon-reload`.
When configuration and data are ready, start explicitly with
`sudo systemctl enable --now keiji` (or `sudo systemctl restart keiji` for an
existing service). Review status with `systemctl status keiji` and logs with
`journalctl -u keiji`. Back up existing data before upgrading or migrating.
