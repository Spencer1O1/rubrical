#!/usr/bin/env python3
"""Block further tools while verification fails. Cloud-agent enforcement path."""

from __future__ import annotations

import json
import sys

from verify_lib import combined_failure_message

# Agent must still read/write to fix failures. Everything else waits.
ALLOWED_WHILE_BLOCKED = frozenset({"Read", "Write"})


def main() -> int:
    payload = json.load(sys.stdin)
    tool_name = payload.get("tool_name", "")
    failure = combined_failure_message()

    if not failure:
        print(json.dumps({"permission": "allow"}))
        return 0

    if tool_name in ALLOWED_WHILE_BLOCKED:
        print(json.dumps({"permission": "allow"}))
        return 0

    print(
        json.dumps(
            {
                "permission": "deny",
                "agent_message": (
                    "Verification failed. Fix these issues before continuing "
                    f"with {tool_name}:\n\n{failure}"
                ),
            }
        )
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
