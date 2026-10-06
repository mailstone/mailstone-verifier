#!/usr/bin/env python3
"""Build the MailStone Verifier application icons from the MailStone app icon.

Usage: make_icons.py <app_icon.png>

<app_icon.png> is the icon the MailStone mobile/desktop application ships
(assets/logo/app_icon.png in MailStoneApp): the black spiral mark on an
opaque white square, 1024×1024. Using the very same artwork keeps every
MailStone application recognisable in a dock, and an opaque image avoids
the alpha-handling differences between icon loaders (an RGBA icon showed up
as a black tile on one Linux desktop).

Writes, next to this script:
  appicon.png          1024×1024 — Wails' source icon (macOS/Windows bundles)
  windows/icon.ico     16…256 px — embedded in the Windows executable
  darwin/iconfile.icns macOS bundle icon, corners rounded as macOS expects
  linux/appicon.png    512 px — window icon (main.go) and hicolor/.desktop entry
"""
import os
import sys

from PIL import Image, ImageDraw

SIZE = 1024


def rounded(img: Image.Image, ratio: float = 0.22) -> Image.Image:
    """macOS-style rounded corners (transparent outside)."""
    mask = Image.new("L", img.size, 0)
    ImageDraw.Draw(mask).rounded_rectangle((0, 0, img.width - 1, img.height - 1), radius=round(img.width * ratio), fill=255)
    out = img.convert("RGBA").copy()
    out.putalpha(mask)
    return out


def main() -> None:
    here = os.path.dirname(os.path.abspath(__file__))
    src = Image.open(sys.argv[1]).convert("RGBA")
    # Flatten on white: the source is opaque already, this only guards a
    # future transparent source.
    master = Image.new("RGBA", src.size, (255, 255, 255, 255))
    master.alpha_composite(src)
    master = master.resize((SIZE, SIZE), Image.LANCZOS)
    master.save(os.path.join(here, "appicon.png"))
    for sub in ("windows", "darwin", "linux"):
        os.makedirs(os.path.join(here, sub), exist_ok=True)
    sizes = [16, 24, 32, 48, 64, 128, 256]
    master.save(os.path.join(here, "windows", "icon.ico"), sizes=[(s, s) for s in sizes])
    rounded(master).save(os.path.join(here, "darwin", "iconfile.icns"))
    master.resize((512, 512), Image.LANCZOS).save(os.path.join(here, "linux", "appicon.png"))
    # Icon theme set for the .desktop entry (GNOME takes the dock icon from
    # here, never from the window): hicolor/<size>x<size>/apps/<app id>.png
    for s in (16, 24, 32, 48, 64, 128, 256, 512):
        d = os.path.join(here, "linux", "icons", "hicolor", f"{s}x{s}", "apps")
        os.makedirs(d, exist_ok=True)
        master.resize((s, s), Image.LANCZOS).save(os.path.join(d, "mailstone-verifier.png"))
    print("icons written: appicon.png, windows/icon.ico, darwin/iconfile.icns, linux/appicon.png, linux/icons/hicolor/*")


if __name__ == "__main__":
    main()
