#!/usr/bin/env python3
"""Run verification after edits. Works on local and cloud agents (afterFileEdit)."""

from __future__ import annotations

import json
import sys

from verify_lib import (
    clear_dirty,
    is_schema_path,
    is_ts_path,
    rel_path,
    set_failure,
    touch_dirty,
    verify_bucket,
)


def main() -> int:
    payload = json.load(sys.stdin)
    rel = rel_path(payload.get("file_path", ""))
    if rel is None:
        return 0

    if is_ts_path(rel):
        touch_dirty("ts_dirty")
        message = verify_bucket("ts")
        set_failure("ts", message)
        if not message:
            clear_dirty("ts_dirty")

    if is_schema_path(rel):
        touch_dirty("schema_dirty")
        message = verify_bucket("schema")
        set_failure("schema", message)
        if not message:
            clear_dirty("schema_dirty")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
