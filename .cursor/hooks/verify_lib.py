"""Shared verification checks for Cursor hooks."""

from __future__ import annotations

import json
import os
import re
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
STATE_DIR = Path(__file__).resolve().parent / "state"
FAILURE_STATE = STATE_DIR / "failures.json"
MIGRATION = ROOT / "migrations" / "00001_initial_schema.sql"
SCHEMA = ROOT / "sql" / "schema" / "schema.sql"
MAX_OUTPUT_CHARS = 6000

TS_SUFFIXES = {".ts", ".tsx"}
TS_FILES = {"extension/tsconfig.json", "extension/package.json"}
SCHEMA_PREFIXES = ("migrations/", "sql/schema/", "sql/queries/")


def rel_path(file_path: str) -> str | None:
    path = Path(file_path)
    try:
        return path.resolve().relative_to(ROOT.resolve()).as_posix()
    except ValueError:
        return None


def is_ts_path(rel: str) -> bool:
    return rel.startswith("extension/") and (
        Path(rel).suffix in TS_SUFFIXES or rel in TS_FILES
    )


def is_schema_path(rel: str) -> bool:
    return rel.startswith(SCHEMA_PREFIXES)


def touch_dirty(name: str) -> None:
    STATE_DIR.mkdir(parents=True, exist_ok=True)
    (STATE_DIR / name).touch()


def dirty(name: str) -> bool:
    return (STATE_DIR / name).is_file()


def clear_dirty(name: str) -> None:
    path = STATE_DIR / name
    if path.is_file():
        path.unlink()


def augment_path(env: dict[str, str]) -> dict[str, str]:
    home = Path.home()
    prefixes: list[str] = []
    nvm = home / ".nvm" / "versions" / "node"
    if nvm.is_dir():
        versions = sorted(nvm.iterdir(), key=lambda p: p.name)
        if versions:
            prefixes.append(str(versions[-1] / "bin"))
    prefixes.extend(
        [
            str(home / ".local" / "share" / "pnpm"),
            "/usr/local/bin",
            "/usr/bin",
        ]
    )
    env = env.copy()
    env["PATH"] = os.pathsep.join(prefixes + [env.get("PATH", "")])
    return env


def run(command: list[str], cwd: Path = ROOT) -> tuple[int, str]:
    env = augment_path(os.environ)
    try:
        completed = subprocess.run(
            command,
            cwd=cwd,
            env=env,
            capture_output=True,
            text=True,
            check=False,
        )
    except FileNotFoundError as exc:
        return 127, str(exc)
    output = (completed.stdout + completed.stderr).strip()
    return completed.returncode, output


def goose_up_sql(text: str) -> str:
    lines: list[str] = []
    for line in text.splitlines():
        if line.strip().startswith("-- +goose Down"):
            break
        lines.append(line)
    return "\n".join(lines)


def normalize_sql(text: str) -> str:
    statements: list[str] = []
    for line in text.splitlines():
        line = line.strip()
        if not line or line.startswith("--"):
            continue
        line = re.sub(r"\bIF NOT EXISTS\b", "", line, flags=re.IGNORECASE)
        line = re.sub(r"\s+", " ", line).strip().rstrip(";")
        if line:
            statements.append(line.lower())
    return "\n".join(sorted(statements))


def schema_files_match() -> str | None:
    if not MIGRATION.is_file() or not SCHEMA.is_file():
        return "Missing migrations/00001_initial_schema.sql or sql/schema/schema.sql."

    migration_sql = normalize_sql(goose_up_sql(MIGRATION.read_text(encoding="utf-8")))
    schema_sql = normalize_sql(SCHEMA.read_text(encoding="utf-8"))
    if migration_sql != schema_sql:
        return (
            "sql/schema/schema.sql does not match migrations/00001_initial_schema.sql "
            "(Up section). Update both together, then squash any incremental migrations."
        )
    return None


def extra_migrations() -> list[str]:
    migrations_dir = ROOT / "migrations"
    if not migrations_dir.is_dir():
        return []
    return sorted(
        path.name
        for path in migrations_dir.glob("*.sql")
        if path.name != "00001_initial_schema.sql"
    )


def check_schema() -> str | None:
    extras = extra_migrations()
    if extras:
        listed = ", ".join(extras)
        return (
            f"Incremental migration files must not remain in git: {listed}. "
            "Squash into migrations/00001_initial_schema.sql and delete them."
        )
    return schema_files_match()


def check_typecheck() -> str | None:
    code, output = run(["pnpm", "--filter", "rubrical-extension", "typecheck"])
    if code == 0:
        return None
    if code == 127:
        code, output = run(["npm", "run", "typecheck"], cwd=ROOT / "extension")
        if code == 0:
            return None
    trimmed = output[-MAX_OUTPUT_CHARS:] if len(output) > MAX_OUTPUT_CHARS else output
    return (
        "Typecheck failed. Fix every reported error before finishing.\n\n"
        f"```\n{trimmed}\n```"
    )


def load_failures() -> dict[str, str]:
    if not FAILURE_STATE.is_file():
        return {}
    try:
        data = json.loads(FAILURE_STATE.read_text(encoding="utf-8"))
    except json.JSONDecodeError:
        return {}
    if not isinstance(data, dict):
        return {}
    return {key: value for key, value in data.items() if isinstance(value, str) and value}


def save_failures(failures: dict[str, str]) -> None:
    STATE_DIR.mkdir(parents=True, exist_ok=True)
    active = {key: value for key, value in failures.items() if value}
    if active:
        FAILURE_STATE.write_text(
            json.dumps(active, indent=2) + "\n", encoding="utf-8"
        )
    elif FAILURE_STATE.is_file():
        FAILURE_STATE.unlink()


def set_failure(bucket: str, message: str | None) -> None:
    failures = load_failures()
    if message:
        failures[bucket] = message
    else:
        failures.pop(bucket, None)
    save_failures(failures)


def combined_failure_message() -> str | None:
    failures = load_failures()
    if not failures:
        return None
    return "\n\n".join(failures.values())


def verify_bucket(bucket: str) -> str | None:
    if bucket == "ts":
        return check_typecheck()
    if bucket == "schema":
        return check_schema()
    raise ValueError(f"unknown bucket: {bucket}")
