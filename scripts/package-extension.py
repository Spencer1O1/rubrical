#!/usr/bin/env python3
"""Build static/downloads/rubrical-extension.zip for /install."""

from __future__ import annotations

import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
EXT = ROOT / "extension"
OUT = ROOT / "static" / "downloads" / "rubrical-extension.zip"
TOP_LEVEL = ("manifest.json", "background.js", "popup.js", "popup.html")


def main() -> None:
    for name in TOP_LEVEL:
        path = EXT / name
        if not path.is_file():
            raise SystemExit(f"missing {path} — run make extension-build-prod first")
    dist = EXT / "dist"
    if not dist.is_dir():
        raise SystemExit(f"missing {dist} — run make extension-build-prod first")

    OUT.parent.mkdir(parents=True, exist_ok=True)
    if OUT.exists():
        OUT.unlink()

    with zipfile.ZipFile(OUT, "w", compression=zipfile.ZIP_DEFLATED) as zf:
        for name in TOP_LEVEL:
            zf.write(EXT / name, name)
        for path in sorted(dist.rglob("*")):
            if path.is_file():
                zf.write(path, path.relative_to(EXT).as_posix())

    print(f"wrote {OUT} ({OUT.stat().st_size} bytes)")


if __name__ == "__main__":
    main()
