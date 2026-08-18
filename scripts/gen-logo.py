"""Generate the [gitplm] logo assets at the repo root.

Renders "[gitplm]" as embedded glyph outlines (no font dependency in the
SVG) using Fira Code Medium, in phosphor green on black like a classic
CRT terminal.

Usage:
    python scripts/gen-logo.py
    rsvg-convert -w 400 gitplm-logo.svg -o gitplm-logo.png
    rsvg-convert -w 360 -h 360 gitplm-logo-square.svg -o gitplm-logo-square.png

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

# Build one combined path, y-flipped (font coords -> SVG coords), advancing x.
pen = SVGPathPen(glyphs)
bpen = BoundsPen(glyphs)
x = 0
for ch in TEXT:
    g = glyphs[cmap[ord(ch)]]
    t = Transform(1, 0, 0, -1, x, 0)
    g.draw(TransformPen(pen, t))
    g.draw(TransformPen(bpen, t))
    x += g.width
path = pen.getCommands()
xmin, ymin, xmax, ymax = bpen.bounds
w, h = xmax - xmin, ymax - ymin


def svg(width, height, scale, tx, ty):
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {width} {height}" width="{width}" height="{height}">
  <rect width="{width}" height="{height}" fill="{BG}"/>
  <g transform="translate({tx:.2f}, {ty:.2f}) scale({scale:.6f})">
    <path d="{path}" fill="{COLOR}"/>
  </g>
</svg>
"""


# Wide wordmark, normalized to 100 px tall.
pad = 0.30 * h
s = 100.0 / (h + 2 * pad)
W = round((w + 2 * pad) * s, 2)
(ROOT / "gitplm-logo.svg").write_text(
    svg(W, 100, s, (pad - xmin) * s, (pad - ymin) * s)
)
print(f"wrote gitplm-logo.svg {W} x 100")

# Square logo: 360x360, text ~85% of width, centered.
SQ = 360
s2 = (SQ * 0.85) / w
tx = (SQ - w * s2) / 2 - xmin * s2
ty = (SQ - h * s2) / 2 - ymin * s2
(ROOT / "gitplm-logo-square.svg").write_text(svg(SQ, SQ, s2, tx, ty))
print("wrote gitplm-logo-square.svg 360x360")
