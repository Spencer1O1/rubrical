#!/usr/bin/env python3
"""Turn-end verification with auto-continue. Local agent only (stop hook)."""

from __future__ import annotations

import json
import sys

from verify_lib import clear_dirty, combined_failure_message, dirty, set_failure, verify_bucket


def main() -> int:
    payload = json.load(sys.stdin)
    if payload.get("status") != "completed":
        print("{}")
        return 0

    if dirty("ts_dirty"):
        message = verify_bucket("ts")
        set_failure("ts", message)
        if not message:
            clear_dirty("ts_dirty")

    if dirty("schema_dirty"):
        message = verify_bucket("schema")
        set_failure("schema", message)
        if not message:
            clear_dirty("schema_dirty")

    failure = combined_failure_message()
    if failure:
        followup = (
            "Verification failed before this task can be considered done:\n\n"
            f"{failure}\n\n"
            "Fix the issues above and re-run the relevant commands."
        )
        print(json.dumps({"followup_message": followup}))
    else:
        print("{}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
