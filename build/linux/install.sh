#!/bin/sh
# Installs MailStone Verifier for the current user: binary, .desktop entry
# and icon theme files. Without the entry, GNOME shows a dark placeholder
# instead of the MailStone logo in the dock and the window switcher.
#
#   build/linux/install.sh            # after `wails build`
#   build/linux/install.sh --remove   # undo
set -eu
here=$(cd "$(dirname "$0")" && pwd)
bin="$here/../bin/mailstone-verifier"
prefix="${XDG_DATA_HOME:-$HOME/.local/share}"
bindir="$HOME/.local/bin"

if [ "${1:-}" = "--remove" ]; then
    rm -f "$bindir/mailstone-verifier" "$prefix/applications/mailstone-verifier.desktop"
    for s in 16 24 32 48 64 128 256 512; do rm -f "$prefix/icons/hicolor/${s}x${s}/apps/mailstone-verifier.png"; done
    command -v update-desktop-database >/dev/null && update-desktop-database "$prefix/applications" || true
    command -v gtk-update-icon-cache >/dev/null && gtk-update-icon-cache -q -t "$prefix/icons/hicolor" || true
    echo "removed"
    exit 0
fi

[ -x "$bin" ] || { echo "binary not found: $bin (run 'wails build' first)" >&2; exit 1; }
install -Dm755 "$bin" "$bindir/mailstone-verifier"
for s in 16 24 32 48 64 128 256 512; do
    install -Dm644 "$here/icons/hicolor/${s}x${s}/apps/mailstone-verifier.png" "$prefix/icons/hicolor/${s}x${s}/apps/mailstone-verifier.png"
done
# Exec must be absolute: ~/.local/bin is not always on the launcher's PATH.
sed "s|^Exec=.*|Exec=$bindir/mailstone-verifier|" "$here/mailstone-verifier.desktop" > "$prefix/applications/mailstone-verifier.desktop"
chmod 644 "$prefix/applications/mailstone-verifier.desktop"
command -v update-desktop-database >/dev/null && update-desktop-database "$prefix/applications" || true
command -v gtk-update-icon-cache >/dev/null && gtk-update-icon-cache -q -t "$prefix/icons/hicolor" || true
echo "installed: $bindir/mailstone-verifier, $prefix/applications/mailstone-verifier.desktop, icons in $prefix/icons/hicolor"
