"""Generate the [gitplm] logo assets at the repo root.

Renders "[gitplm]" as embedded glyph outlines (no font dependency in the
SVG) using Fira Code Medium, in phosphor green on black like a classic
CRT terminal.

Outputs:
    gitplm-logo.svg         wide wordmark (README header)
    gitplm-logo-square.svg  square wordmark (og:image / social previews)
    gitplm-icon.svg         square [g] icon (favicons, small sizes)

Usage:
    python scripts/gen-logo.py
    rsvg-convert -w 400 gitplm-logo.svg -o gitplm-logo.png
    rsvg-convert -w 360 -h 360 gitplm-logo-square.svg -o gitplm-logo-square.png
    rsvg-convert -w 64 -h 64 gitplm-icon.svg -o favicon.png

Requires fontTools and the Fira Code font (path below).
"""
from pathlib import Path

from fontTools.misc.transform import Transform
from fontTools.pens.boundsPen import BoundsPen
from fontTools.pens.svgPathPen import SVGPathPen
from fontTools.pens.transformPen import TransformPen
from fontTools.ttLib import TTFont

FONT = "/usr/share/fonts/TTF/FiraCodeNerdFont-Medium.ttf"
TEXT = "[gitplm]"
COLOR = "#33ff33"  # phosphor green
BG = "#000000"

ROOT = Path(__file__).resolve().parent.parent

font = TTFont(FONT)
cmap = font.getBestCmap()
glyphs = font.getGlyphSet()


def text_path(text):
    """Combined outline path, y-flipped (font coords -> SVG coords), plus bounds."""
    pen = SVGPathPen(glyphs)
    bpen = BoundsPen(glyphs)
    x = 0
    for ch in text:
        g = glyphs[cmap[ord(ch)]]
        t = Transform(1, 0, 0, -1, x, 0)
        g.draw(TransformPen(pen, t))
        g.draw(TransformPen(bpen, t))
        x += g.width
    return pen.getCommands(), bpen.bounds


def svg(path, width, height, scale, tx, ty):
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {width} {height}" width="{width}" height="{height}">
  <rect width="{width}" height="{height}" fill="{BG}"/>
  <g transform="translate({tx:.2f}, {ty:.2f}) scale({scale:.6f})">
    <path d="{path}" fill="{COLOR}"/>
  </g>
</svg>
"""


def square(path, bounds, frac):
    """360x360 square with the text scaled to frac of the width, centered."""
    xmin, ymin, xmax, ymax = bounds
    w, h = xmax - xmin, ymax - ymin
    SQ = 360
    s = (SQ * frac) / w
    tx = (SQ - w * s) / 2 - xmin * s
    ty = (SQ - h * s) / 2 - ymin * s
    return svg(path, SQ, SQ, s, tx, ty)


path, bounds = text_path(TEXT)
xmin, ymin, xmax, ymax = bounds
w, h = xmax - xmin, ymax - ymin

# Wide wordmark, normalized to 100 px tall.
pad = 0.30 * h
s = 100.0 / (h + 2 * pad)
W = round((w + 2 * pad) * s, 2)
(ROOT / "gitplm-logo.svg").write_text(
    svg(path, W, 100, s, (pad - xmin) * s, (pad - ymin) * s)
)
print(f"wrote gitplm-logo.svg {W} x 100")

# Square logo: 360x360, text ~85% of width, centered.
(ROOT / "gitplm-logo-square.svg").write_text(square(path, bounds, 0.85))
print("wrote gitplm-logo-square.svg 360x360")

# [g] icon for favicons -- a full wordmark is unreadable at tab size.
ipath, ibounds = text_path("[g]")
(ROOT / "gitplm-icon.svg").write_text(square(ipath, ibounds, 0.72))
print("wrote gitplm-icon.svg 360x360")
