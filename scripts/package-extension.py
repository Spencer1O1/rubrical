#!/usr/bin/env python3
"""Build static/downloads/rubrical-extension.zip for /install."""

from __future__ import annotations

import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
EXT = ROOT / "extension"
OUT = ROOT / "static" / "downloads" / "rubrical-extension.zip"
ROOT_FILES = ("manifest.json", "popup.html")
DIST_FILES = ("content.js", "background.js", "popup.js")


def main() -> None:
    for name in ROOT_FILES:
        path = EXT / name
        if not path.is_file():
            raise SystemExit(f"missing {path}")
    dist = EXT / "dist"
    for name in DIST_FILES:
        path = dist / name
        if not path.is_file():
            raise SystemExit(f"missing {path} — run make extension-build-prod first")

    OUT.parent.mkdir(parents=True, exist_ok=True)
    if OUT.exists():
        OUT.unlink()

    with zipfile.ZipFile(OUT, "w", compression=zipfile.ZIP_DEFLATED) as zf:
        for name in ROOT_FILES:
            zf.write(EXT / name, name)
        for name in DIST_FILES:
            path = dist / name
            zf.write(path, f"dist/{name}")

    print(f"wrote {OUT} ({OUT.stat().st_size} bytes)")


if __name__ == "__main__":
    main()
