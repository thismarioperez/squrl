#!/usr/bin/env python3
"""
gen-cli-icon.py — Convert a pixel-art PNG to ANSI half-block terminal art.

Each pair of vertical pixels maps to one terminal row using ▄ (lower half-block):
  - top pixel → background colour
  - bottom pixel → foreground colour

Transparent pixels are treated as terminal background (no colour code emitted).
Colours are written as exact 24-bit ANSI sequences so no blending occurs.

Usage: gen-cli-icon.py input.png > output.txt
"""

import sys
import zlib
import struct


def _defilter_row(filter_type, raw, prev, bpp):
    out = bytearray(len(raw))
    for i, b in enumerate(raw):
        a = out[i - bpp] if i >= bpp else 0
        c = prev[i - bpp] if i >= bpp else 0
        p = prev[i]
        if filter_type == 0:
            out[i] = b
        elif filter_type == 1:
            out[i] = (b + a) & 0xFF
        elif filter_type == 2:
            out[i] = (b + p) & 0xFF
        elif filter_type == 3:
            out[i] = (b + ((a + p) >> 1)) & 0xFF
        elif filter_type == 4:
            pa = abs(p - c)
            pb = abs(a - c)
            pc = abs(a + p - 2 * c)
            pr = a if pa <= pb and pa <= pc else (p if pb <= pc else c)
            out[i] = (b + pr) & 0xFF
    return out


def read_png(path):
    with open(path, "rb") as f:
        data = f.read()
    assert data[:8] == b"\x89PNG\r\n\x1a\n", "not a PNG file"

    pos = 8
    ihdr = None
    idat_chunks = []
    while pos < len(data):
        length = struct.unpack(">I", data[pos : pos + 4])[0]
        chunk_type = data[pos + 4 : pos + 8]
        chunk_data = data[pos + 8 : pos + 8 + length]
        if chunk_type == b"IHDR":
            ihdr = chunk_data
        elif chunk_type == b"IDAT":
            idat_chunks.append(chunk_data)
        pos += 12 + length

    width, height = struct.unpack(">II", ihdr[:8])
    bit_depth, color_type = ihdr[8], ihdr[9]
    assert bit_depth == 8, f"only 8-bit PNGs supported (got {bit_depth})"
    assert color_type == 6, f"only RGBA PNGs supported (color_type={color_type})"

    raw = zlib.decompress(b"".join(idat_chunks))
    stride = width * 4  # RGBA = 4 bytes per pixel
    prev = bytearray(stride)
    pixels = []
    for y in range(height):
        row_start = y * (stride + 1)
        filter_type = raw[row_start]
        row = bytearray(raw[row_start + 1 : row_start + 1 + stride])
        row = _defilter_row(filter_type, row, prev, 4)
        prev = row
        row_pixels = []
        for x in range(width):
            base = x * 4
            row_pixels.append(tuple(row[base : base + 4]))  # (R, G, B, A)
        pixels.append(row_pixels)

    return width, height, pixels


def main():
    if len(sys.argv) != 2:
        print(f"usage: {sys.argv[0]} input.png", file=sys.stderr)
        sys.exit(1)

    width, height, pixels = read_png(sys.argv[1])
    assert height % 2 == 0, "image height must be even for half-block mapping"

    ESC = "\033"
    RESET = f"{ESC}[0m"
    HIDE = f"{ESC}[?25l"
    SHOW = f"{ESC}[?25h"

    def fg(r, g, b):
        return f"{ESC}[38;2;{r};{g};{b}m"

    def bg(r, g, b):
        return f"{ESC}[48;2;{r};{g};{b}m"

    lines = []
    for row in range(0, height, 2):
        line = ""
        for col in range(width):
            tr, tg, tb, ta = pixels[row][col]
            br, bg_, bb, ba = pixels[row + 1][col]

            if ta == 0 and ba == 0:
                line += RESET + " "
            elif ta == 0:
                line += RESET + fg(br, bg_, bb) + "▄"
            elif ba == 0:
                line += RESET + fg(tr, tg, tb) + "▀"
            else:
                line += RESET + bg(tr, tg, tb) + fg(br, bg_, bb) + "▄"
        line += RESET
        lines.append(line)

    sys.stdout.write(HIDE)
    sys.stdout.write("\n".join(lines))
    sys.stdout.write("\n")
    sys.stdout.write(SHOW)


if __name__ == "__main__":
    main()
