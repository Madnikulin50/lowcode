#!/bin/sh
# Converts a DWG to PDF for stroykontrol (see dwg.go): dwg2SVG (LibreDWG) -> rsvg-convert.
# Text stays text in the PDF, so pdftotext / the title-block reader work on it.
# usage: dwg2pdf.sh in.dwg out.pdf
# To use another converter (e.g. a commercial CAD CLI) pass --dwg-converter
# "<cmd> {in} {out}" instead of this script.
set -eu
[ $# -eq 2 ] || { echo "usage: $0 in.dwg out.pdf" >&2; exit 2; }
svg="$(mktemp --suffix=.svg)"
trap 'rm -f "$svg"' EXIT
dwg2SVG "$1" > "$svg"
[ -s "$svg" ] || { echo "dwg2SVG produced no output for $1" >&2; exit 1; }
# dwg2SVG writes width/height="100%" and drawing units in viewBox, which makes
# rsvg-convert emit a page of thousands of points. Pin the long side to 2384 pt (A1).
dims="$(sed -n 's/.*viewBox="\([^"]*\)".*/\1/p' "$svg" | head -n1 | awk '{
  w=$3; h=$4; if (w<=0 || h<=0) exit 1
  k=2384/(w>h?w:h); printf "%d %d", w*k+1, h*k+1 }')" || dims=""
if [ -n "$dims" ]; then
  set -- "$1" "$2" "${dims% *}" "${dims#* }"
  sed -i '0,/width="100%" height="100%"/s//width="'"$3"'" height="'"$4"'"/' "$svg"
fi
rsvg-convert -f pdf -o "$2" "$svg"
