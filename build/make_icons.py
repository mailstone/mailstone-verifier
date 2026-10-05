#!/usr/bin/env python3
"""Build the MailStone Verifier application icons from the MailStone mark.

Usage: make_icons.py <mark.png>

<mark.png> is the spiral mark, dark on transparent (any size; the only
rasters in the repositories are 176 px, so the mark is upscaled with a
threshold on its alpha: the glyph is solid shapes, and re-binarising the
upscaled alpha gives crisp edges instead of a blurred enlargement).

Writes, next to this script:
  appicon.png          1024×1024 — Wails' source icon (macOS/Windows bundles)
  windows/icon.ico     16…256 px — embedded in the Windows executable
  darwin/iconfile.icns macOS bundle icon
  linux/appicon.png    512 px — window icon + .desktop entry on Linux

Charter: white mark on the MailStone navy (#1f2b45), orange (#e28638)
hairline at the bottom, 22 % corner radius.
"""
import os
import sys

from PIL import Image, ImageDraw, ImageFilter

NAVY = (31, 43, 69, 255)
ORANGE = (226, 134, 56, 255)
SIZE = 1024


def mark_mask(src: Image.Image, target: int) -> Image.Image:
    """White glyph mask of the mark at `target` px, edges re-sharpened."""
    src = src.convert("RGBA")
    # The mark is dark on transparent (or dark on white): use alpha when the
    # file has one, else darkness.
    alpha = src.getchannel("A")
    if alpha.getextrema() == (255, 255):
        alpha = src.convert("L").point(lambda v: 255 - v)
    bbox = alpha.getbbox()
    glyph = alpha.crop(bbox)
    w, h = glyph.size
    scale = target / max(w, h)
    big = glyph.resize((round(w * scale), round(h * scale)), Image.LANCZOS)
    # Binarise, then soften by one pixel so the outline is not aliased.
    big = big.point(lambda v: 255 if v > 128 else 0).filter(ImageFilter.GaussianBlur(0.8))
    return big


def compose(mask: Image.Image, size: int) -> Image.Image:
    icon = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    draw = ImageDraw.Draw(icon)
    radius = round(size * 0.22)
    draw.rounded_rectangle((0, 0, size - 1, size - 1), radius=radius, fill=NAVY)
    # Orange hairline inside the bottom edge, like the app's title bar.
    band = round(size * 0.035)
    strip = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    ImageDraw.Draw(strip).rounded_rectangle((0, 0, size - 1, size - 1), radius=radius, fill=ORANGE)
    cut = Image.new("L", (size, size), 0)
    ImageDraw.Draw(cut).rectangle((0, size - band, size, size), fill=255)
    icon.paste(strip, (0, 0), cut)
    # The mark, white, centred, 64 % of the tile.
    factor = size * 0.64 / max(mask.size)
    glyph = mask.resize((round(mask.width * factor), round(mask.height * factor)), Image.LANCZOS)
    white = Image.new("RGBA", glyph.size, (255, 255, 255, 255))
    x = (size - glyph.width) // 2
    y = (size - band - glyph.height) // 2
    icon.paste(white, (x, y), glyph)
    return icon


def main() -> None:
    here = os.path.dirname(os.path.abspath(__file__))
    src = Image.open(sys.argv[1])
    mask = mark_mask(src, SIZE)
    master = compose(mask, SIZE)
    master.save(os.path.join(here, "appicon.png"))
    for sub in ("windows", "darwin", "linux"):
        os.makedirs(os.path.join(here, sub), exist_ok=True)
    sizes = [16, 24, 32, 48, 64, 128, 256]
    master.save(os.path.join(here, "windows", "icon.ico"), sizes=[(s, s) for s in sizes])
    master.save(os.path.join(here, "darwin", "iconfile.icns"))
    master.resize((512, 512), Image.LANCZOS).save(os.path.join(here, "linux", "appicon.png"))
    print("icons written: appicon.png, windows/icon.ico, darwin/iconfile.icns, linux/appicon.png")


if __name__ == "__main__":
    main()
