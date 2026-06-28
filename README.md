# Rubrical

> Check the rubric before the rubric checks you.

A rubric-aware preflight checker for student assignments — import Canvas assignment context, review your draft against the rubric, and get structured feedback before you submit.

## Documentation

| Doc | Description |
|-----|-------------|
| [Development guide](docs/development.md) | Setup, WSL, Docker, Makefile |
| [Specification](docs/specification.md) | Full product & technical spec |
| [Spec checklist](docs/spec-checklist.md) | MVP progress, gaps, and what's next |

## Quick start

```bash
pnpm install && make setup
make db-up && make migrate-up
make css-watch    # terminal 1
make templ-watch  # terminal 2
make server       # terminal 3
make extension-build
```

Open http://localhost:8787 — see [docs/development.md](docs/development.md) for details.
