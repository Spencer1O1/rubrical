# Canvas page fixtures

Saved HTML snapshots from live Canvas for extension DOM extraction tests.

## Layout

| Path | Role |
|------|------|
| `*.html` | Standalone fixture pages (cleaned via `pnpm clean:fixtures`) |
| `expectations/<stem>.json` | Contract per case — anchors, metadata extraction, optional `compose` |
| `fragments/*.html` | DOM snippets swapped onto a base page for variants |

**Composed cases** (no standalone html — built at test time from base + fragment):

- `assignment-file-uploaded` — `assignment-file-upload.html` + upload pane with a completed file row

## Verified selectors (in use today)

Canvas DOM anchors live under [`extension/src/canvas/anchors/`](../extension/src/canvas/anchors/). Each module documents **a2** (verified `data-testid`), **classic**, and **extra** tiers. Lookups go through [`extension/src/canvas/query.ts`](../extension/src/canvas/query.ts) (`queryAnchor`, `firstMatch`, etc.).

Dynamic Canvas ids (e.g. `traditional-criterion-_9691-ratings-0`) use wildcard helpers: `testIdStartsWith("traditional-criterion-")`, `testIdContains("-ratings-")` — not literal strings.

| Area | Anchor module | Key `data-testid` values |
|------|---------------|--------------------------|
| File upload | [`upload.ts`](../extension/src/canvas/anchors/upload.ts) | `uploaded_files_table`, `input-file-drop`, `upload-pane`, `attempt-tab`, `assignment-2-student-content-tabs` |
| Submission tabs / draft | [`submission.ts`](../extension/src/canvas/anchors/submission.ts) | `submission-type-selector`, `online_upload`, `online_url`, `online_text_entry`, `text-editor`, `url-input` |
| Instructions | [`instructions.ts`](../extension/src/canvas/anchors/instructions.ts) | `assignments-2-assignment-description`, `assignments-2-student-view` |
| Rubric | [`rubric.ts`](../extension/src/canvas/anchors/rubric.ts) | `traditional-view-criterion-ratings`, `rubric-assessment-traditional-view`, `long-description-close-button`; per-rating cells via `traditional-criterion-*-ratings-*` wildcards |
| Submit / reply | [`submit.ts`](../extension/src/canvas/anchors/submit.ts) | `submit-button` (assignments), `DiscussionEdit-submit` (open discussion composer), `discussion-topic-reply` (closed — not used for placement) |
| Discussion | [`discussion.ts`](../extension/src/canvas/anchors/discussion.ts) | `discussion-post-menu-trigger`, `discussion-thread-menuitem-rubric`, `DiscussionEdit-container`, `message-body`, `attach-btn`, `attachment-input`, `removable-item` (staged attachment), `assignment-rubric-modal`, `preview-assignment-rubric-button`, `enhanced-rubric-assessment-tray` |
| Metadata | [`metadata/`](../extension/src/canvas/anchors/metadata/) | `due-date` (`datetime` ISO attr), `grade-display`; submission type derived from `submission.ts` type blocks |

Trash `button[id="<digits>"]` per upload row is defined in `upload.ts` (`fileRow.trashButton`).

`online_text_entry` is the submission **type tab** control. The mounted rich-text editor lives under `text-editor` when that tab is active — see `assignment-text-tab.html`.

Graded discussion rubrics open via **post menu (`discussion-post-menu-trigger`) → Show Rubric (`discussion-thread-menuitem-rubric`) → Preview Rubric** — not the points label.

Do **not** add new selectors without a fixture row here and an anchor module entry. Do not use CSS module class names (`css-*`) as selectors.

## Fixture contract tests

Extension tests load each expectations case and compare against `expectations/<stem>.json`:

- **`anchors`** — whether `queryAnchor` / `firstMatch` finds each keyed anchor on that page
- **`extraction`** — metadata outputs (`dueAt`, `dueDateText`, `pointsPossibleText`, `submissionTypeText`, `allowedSubmissionTypes`)
- **`compose`** (optional) — `{ base, replace: [{ selector, fragment }] }` to build variants without duplicating full pages

Run from repo root:

```bash
pnpm clean:fixtures
pnpm test:fixtures
```

When cleaning live Canvas captures, pass your PII via flags, env vars, or prompts (never commit name/email to the repo):

```bash
pnpm clean:fixtures -- --name "Your Name" --email "you@school.edu" --inst-host "school.instructure.com"
# or
RUBRICAL_FIXTURE_DISPLAY_NAME="..." RUBRICAL_FIXTURE_EMAIL="..." pnpm clean:fixtures
```

**Adding a fixture:** save a live Canvas snapshot to `fixtures/my-case.html`, add `fixtures/expectations/my-case.json` (copy a similar case), run `pnpm clean:fixtures`, then `pnpm test:fixtures` and fix expectations until green. Without expectations yet, `clean:fixtures` keeps every `[data-testid]` subtree.

For a variant of an existing page, add a fragment under `fragments/` and an expectations case with `compose` instead of copying the whole html.

## Fixtures

| Case | What it covers |
|------|----------------|
| `assignment-rubric` | Text + URL submission types, rubric, due date, points |
| `assignment-rubric-modal-open` | Same + rubric long-description modal open (InstUI scroll lock on `<html>`) |
| `assignment-file-upload` | Upload tab selected, empty drop zone (no file row yet) |
| `assignment-file-uploaded` | Upload tab with one completed file row (composed) |
| `assignment-text-tab` | Text tab selected with mounted RCE (`text-editor`, `textarea#textentry_text`, `.tox-edit-area__iframe`) |
| `assignment-url-tab` | Web URL tab selected, `url-input` visible |
| `assignment-submitted` | Graded submission read-only view (no submit button until student clicks New Attempt) |
| `discussion-prompt` | Discussion topic prompt + closed Reply opener (`discussion-topic-reply`) |
| `discussion-menu-open` | Post ⋮ menu open with **Show Rubric** (`discussion-thread-menuitem-rubric`) |
| `discussion-reply-open` | Main topic reply composer open (`DiscussionEdit-container`, `DiscussionEdit-submit`, `message-body`) |
| `discussion-rubric-modal` | Assignment Rubric Details modal (`assignment-rubric-modal`, `preview-assignment-rubric-button`) |
| `discussion-rubric-tray` | Rubric assessment tray open (`enhanced-rubric-assessment-tray`, traditional criterion ratings) |
| `discussion-attachment` | Reply composer with one attached file (`removable-item`, download link under `DiscussionEdit-container`) |

Discussion replies are **text + optional single attachment** — no submission-type tabs. The hidden file picker is `attachment-input`; Rubrical stages bytes on pick (same as assignment uploads) and sends them on import. Canvas download URLs are a fallback only after upload completes.

After capture, document any **new** `data-testid` values in the anchor module and table above.
