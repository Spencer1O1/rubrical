# Canvas page fixtures

Saved HTML snapshots from live Canvas for extension DOM extraction tests.

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

`online_text_entry` is the submission **type tab** control. The mounted rich-text editor lives under `text-editor` when that tab is active — see `1-text-submission-tab.html`.

Graded discussion rubrics open via **post menu (`discussion-post-menu-trigger`) → Show Rubric (`discussion-thread-menuitem-rubric`) → Preview Rubric** — not the points label.

Do **not** add new selectors without a fixture row here and an anchor module entry. Do not use CSS module class names (`css-*`) as selectors.

## Fixture contract tests

Extension tests load each `*.html` snapshot and compare against `expectations/<stem>.json`:

- **`anchors`** — whether `queryAnchor` / `firstMatch` finds each keyed anchor on that page
- **`extraction`** — metadata outputs (`dueAt`, `dueDateText`, `pointsPossibleText`, `submissionTypeText`, `allowedSubmissionTypes`)

Run from repo root:

```bash
pnpm test:fixtures
```

A meta-test also asserts every id in `uploadIds`, `submissionIds`, `instructionsIds`, `metadataIds`, `rubricIds`, and `submitIds` appears in at least one HTML fixture.

**Adding a fixture:** save `fixtures/my-case.html`, add `fixtures/expectations/my-case.json`, run tests, fix expectations until green.

## Fixtures

| File | What it covers |
|------|----------------|
| `1-modal-closed.html` | Text + URL submission types, rubric, due date, points |
| `1-modal-open.html` | Same + rubric long-description modal open |
| `1-file-uploaded.html` | File uploaded (`uploaded_files_table` row, numeric trash `button[id]`, `upload-box` / `upload-pane`) |
| `1-text-submission-tab.html` | Text tab selected with mounted RCE (`text-editor`, `textarea#textentry_text`, `.tox-edit-area__iframe`) |
| `2-text-submission.html` | File upload tab selected (`online_upload`), empty upload UI |
| `2-url-submission.html` | Web URL tab selected, `url-input` visible |
| `3-discussion.html` | Discussion topic prompt + closed Reply opener (`discussion-topic-reply`) |
| `3-discussion-three-dots-open.html` | Post ⋮ menu open with **Show Rubric** (`discussion-thread-menuitem-rubric`) |
| `3-discussion-reply-open.html` | Main topic reply composer open (`DiscussionEdit-container`, `DiscussionEdit-submit`, `message-body`) |
| `3-discussion-modal-open.html` | Assignment Rubric Details modal (`assignment-rubric-modal`, `preview-assignment-rubric-button`) |
| `3-discussion-rubric-open.html` | Rubric assessment tray open (`enhanced-rubric-assessment-tray`, traditional criterion ratings) |
| `3-discussion-attachment.html` | Reply composer with one attached file (`removable-item`, download link under `DiscussionEdit-container`) |

Discussion replies are **text + optional single attachment** — no submission-type tabs. The hidden file picker is `attachment-input`; Rubrical stages bytes on pick (same as assignment uploads) and sends them on import. Canvas download URLs are a fallback only after upload completes.

## Gaps

1. **`file-upload-in-progress.html`** — row visible while Canvas is still uploading (progress UI visible).

After capture, document any **new** `data-testid` values in the anchor module and table above.
