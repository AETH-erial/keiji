#!/bin/sh
# Provision Keiji's service identity and private directories on Linux.
set -eu

if [ "$(id -u)" -ne 0 ]; then
    echo 'Run this script as root (for example, sudo ./deploy/setup.sh).' >&2
    exit 1
fi
if [ ! -x /sbin/nologin ]; then
    echo '/sbin/nologin is not available; refusing to create the service user.' >&2
    exit 1
fi

# Do not silently repurpose an existing human account or change its shell.
if getent group keiji >/dev/null; then
    :
else
    groupadd --system keiji
fi
if getent passwd keiji >/dev/null; then
    entry=$(getent passwd keiji)
    existing_gid=$(printf '%s\n' "$entry" | cut -d: -f4)
    existing_shell=$(printf '%s\n' "$entry" | cut -d: -f7)
    group_gid=$(getent group keiji | cut -d: -f3)
    if [ "$existing_shell" != /sbin/nologin ] || [ "$existing_gid" != "$group_gid" ]; then
        echo 'Existing keiji account has a different shell or primary group; review it manually.' >&2
        exit 1
    fi
else
    useradd --system --no-create-home --gid keiji --shell /sbin/nologin keiji
fi

# Refuse symlinks at managed paths rather than following them as root.
for directory in /etc/keiji /var/lib/keiji /var/lib/keiji/images; do
    if [ -L "$directory" ] || { [ -e "$directory" ] && [ ! -d "$directory" ]; }; then
        echo "Not a safe directory: $directory" >&2
        exit 1
    fi
done

# Config is root-managed, readable by the service group, never service-writable.
install -d -o root -g keiji -m 0750 /etc/keiji
# Existing configuration contents are deliberately not changed recursively.
# An administrator must install individual secrets as root:keiji mode 0640.

# SQLite needs write access to its directory for journal and sidecar files.
install -d -o keiji -g root -m 0700 /var/lib/keiji /var/lib/keiji/images
# Do not recursively change ownership or modes on existing application data.

cat <<'MSG'
Keiji identity and directories are ready.
Install your environment file as root:keiji mode 0640, for example:
  install -o root -g keiji -m 0640 /path/to/keiji.env /etc/keiji/keiji.env
Existing files were not modified. Review their ownership, permissions and ACLs.
This script does not install or start Keiji.
MSG
