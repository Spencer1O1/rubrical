# Spec checklist

Track implementation against [specification.md](./specification.md) §14 (MVP build order). Update this file when behavior changes.

**Legend:** ✅ done · 🟡 partial · ❌ not started · 🔴 gap (spec requires, not built)

Last reviewed: 2026-06-26

## Summary

| Area | Status |
|------|--------|
| Phases 1–6 (import shell) | 🟡 Mostly complete |
| Phase 7 (AI analysis) | ❌ Not started — **core product gap** |
| Phase 8 (revision loop) | ❌ Not started |
| Phase 9 (polish) | 🟡 Partial |
| File / external-tool submissions | 🟡 Canvas file read + manual upload; LTI still manual-only |
| Discussions | 🔴 Page detected; prompt + topic reply not verified (§5.6 MVP) |
| Real user auth | 🔴 Local dev user only |

---

## Phase 1: Skeleton

- [x] Go project + chi router
- [x] templ + Tailwind + static assets
- [x] Dashboard page
- [x] Postgres (Docker Compose)
- [x] goose migrations
- [x] sqlc type-safe queries
- [x] Health route (`/health`)
- [x] CORS for extension
- [x] Local dev user (`local@rubrical.dev`)

---

## Phase 2: Browser extension injection

- [x] Manifest V3 extension
- [x] Content script on `*.instructure.com`
- [x] Assignment page detection
- [x] Discussion page detection (page type only)
- [x] MutationObserver + debounced re-injection
- [x] “Check with Rubrical” button near submit controls (single label — same action regardless of draft/submission type)
- [x] Duplicate injection prevention
- [x] No network traffic before user click
- [x] Modal iframe (`?embed=1`) after import

---

## Phase 3: Assignment import

- [x] Extract title, instructions (HTML), visible text
- [x] Extract course name, submission type label (text)
- [x] POST `/imports` JSON payload
- [x] Upsert assignment snapshot by `(user_id, source_url)`
- [x] Redirect to assignment page / modal
- [x] Assignment detail view (dashboard + embed)
- [x] Instructions rendered as sanitized HTML
- [x] Instructions scroll panel (dashboard) + table horizontal scroll wrappers
- [x] Parse/store due date (`dueDateText` → `due_at`, best-effort)
- [x] Parse/store points possible (`pointsPossibleText` → `points_possible`)
- [x] Import payload validation (Canvas URL, size limits, rubric shape, draft file base64)

**Discussions (MVP — see spec §5.6):**

- [x] `page_type: discussion` on `/discussion_topics/`
- [x] Button anchor selectors include discussion reply controls
- [ ] 🔴 **Discussion prompt extraction** — not verified on real Canvas discussion DOM
- [ ] 🔴 **Main topic reply editor** — draft read not verified end-to-end
- [ ] ❌ Thread reply composers + `draftEditorRole: thread_reply` (v2)
- [ ] ❌ AI: participation rules vs main-post criteria (Phase 7)

---

## Phase 4: Draft capture (Canvas text)

- [x] Textarea submission fields
- [x] TinyMCE / iframe editor read order
- [x] Live sync before read (blur + triggerSave)
- [x] Store draft on import when non-empty
- [x] `source_type` + `captured_from_canvas` on draft rows
- [ ] 🔴 Discussion **topic reply** editor — same draft pipeline as assignments; not verified (spec §5.6 MVP)
- [ ] ❌ Discussion **thread reply** editors — one button per composer (v2)
- [ ] 🟡 Previous submission preview text — not explicitly targeted
- [ ] 🟡 Extraction confidence metadata — not sent

---

## Phase 5: Rubric extraction

- [x] Canvas A2 traditional view (primary)
- [x] Classic rubric table fallback
- [x] Strict extraction mode (`RUBRICAL_STRICT_EXTRACTION`)
- [x] Structured rubric: header + rows + ratings (title, description, points)
- [x] Store in `rubric_criteria.ratings_json` + `extracted_sources` header
- [x] Display rubric table (dashboard full-width; embed collapsible)
- [x] Rating column equal width + points bottom-right in cells
- [ ] 🟡 Variable rating column counts across rows — padded via `RatingColumnCount()`
- [ ] 🟡 i18n rubric headers — use Canvas header when present; strict shows `—` when missing

---

## Phase 6: Draft input (Rubrical UI)

- [x] Draft textarea on assignment page (dashboard + embed)
- [x] Word count / draft status label
- [x] `POST /assignments/{id}/draft` save
- [x] HTMX partial on save (`DraftSaved`)
- [x] Analyze button (form post)
- [ ] 🟡 Analyze saves draft then shows placeholder only (no real analysis)
- [x] Paste/upload prompt when upload-type assignment has no draft
- [x] Manual file upload UI + `POST /assignments/{id}/draft/upload`
- [x] Show submission type on assignment screen

---

## Phase 7: AI analysis

**Status: ❌ Not started — blocks core value proposition**

Schema and routes exist; handlers return placeholder only.

- [ ] AI provider interface (Go)
- [ ] OpenAI or Anthropic adapter
- [ ] `OPENAI_API_KEY` / `ANTHROPIC_API_KEY` wired
- [ ] Analysis prompt (instructions + rubric + draft; discussion: full prompt + `draftEditorRole` when present)
- [ ] Structured JSON output per spec §11
- [ ] Validate + persist `analysis_runs`
- [ ] Persist `feedback_items`
- [ ] Render feedback cards (criterion-level, missing requirements, strengths)
- [ ] Estimated score + confidence + evidence
- [ ] Replace `AnalysisPending` placeholder
- [ ] Run analysis on import when draft present (spec §5.3 step 9) — optional policy TBD

---

## Phase 8: Revision loop

- [ ] Re-analyze after draft edit
- [ ] Analysis run history
- [ ] Mark feedback resolved (`POST /feedback/{id}/resolve` is stub)
- [ ] Compare previous vs latest results

---

## Phase 9: Polish

- [x] A2 rubric selectors (iterative)
- [x] Assignment layout: scrollable instructions, full-width rubric (dashboard)
- [x] Embed: collapsed instructions + rubric
- [x] `.env.local` loading
- [x] WSL / extension dev docs
- [ ] Manual extraction fallback UI (region select, etc.)
- [ ] HTMX loading states on Analyze
- [ ] Error handling / user-facing import failures beyond alert
- [ ] Dashboard status badges (`analyzed` vs `imported`)
- [ ] Privacy / settings page
- [ ] Side panel mode (future in spec)

---

## File submissions & external tools

**Status: 🟡 Canvas file read + manual upload built; external-tool/LTI still manual-only**

Spec §5.5 and §5.4 require: when Canvas has no readable draft (file upload, external tool, LTI), import assignment context and let the student **paste or upload** their work in Rubrical.

### What works today

| Capability | Status |
|------------|--------|
| Import assignment + rubric without draft | ✅ |
| Capture Canvas **online text entry** draft | ✅ |
| Read Canvas file input / attachment link | ✅ |
| **IndexedDB staging on Canvas upload/delete (page-origin, content script)** | ✅ |
| **Page-load draft manifest GET (metadata)** | ✅ |
| **Per-row re-upload indicator when file not readable** | ✅ |
| **File-upload tab HTML fixtures** (`fixtures/`) | 🟡 capture in-progress + complete |
| **Import with `draftFileRefs` (no re-upload)** | ✅ |
| **Persist submission file bytes** (`RUBRICAL_DATA_DIR`) | ✅ |
| Manual file upload (assignment + embed) | ✅ |
| Remove / replace stored submission file in UI | ✅ |
| Text / File / Web URL tab switcher (mutually exclusive) | ✅ |
| Tabs filtered to assignment-allowed submission types | ✅ |
| Canvas Web URL capture on import | ✅ |
| Show filename + upload time when file stored | ✅ |
| `body` for Canvas text / manual paste only (no file text extraction) | ✅ |
| Show `submission_type` on assignment screen | ✅ |
| Paste/upload prompt when upload-type + empty draft | ✅ |
| `source_type` on drafts (canvas text/file, manual paste, file upload) | ✅ |

### Remaining gaps

| Capability | Status | Notes |
|------------|--------|--------|
| External tool (LTI) on Canvas | 🟡 | Manual upload in Rubrical; no Canvas draft to read |

---

## Auth & multi-user

- [x] `users` table
- [x] Local dev user bootstrap
- [ ] 🔴 Real login / session
- [ ] 🔴 Per-user assignment isolation in production
- [ ] OAuth / institution SSO (out of v1 scope)

---

## Data model vs spec

| Table | Schema | Used |
|-------|--------|------|
| `assignment_snapshots` | ✅ | ✅ |
| `rubric_criteria` | ✅ | ✅ |
| `submission_drafts` | ✅ (text/url mode) | ✅ `body` + mode |
| `submission_draft_files` | ✅ multi-file | ✅ stored blobs |
| `analysis_runs` | ✅ | ❌ empty |
| `feedback_items` | ✅ | ❌ empty |
| `extracted_sources` | ✅ (migration) | ✅ rubric header |

---

## API routes

| Route | Spec | Status |
|-------|------|--------|
| `GET /` | Dashboard | ✅ |
| `GET /health` | — | ✅ (+ strictExtraction flag) |
| `POST /imports` | Import | ✅ (+ `draftFileRefs`) |
| `GET /assignments/draft-manifest` | Extension metadata | ✅ |
| `GET /assignments/{id}` | Detail | ✅ (+ embed query) |
| `POST /assignments/{id}/draft` | Save draft | ✅ text only |
| `POST /assignments/{id}/draft/upload` | — | ✅ multipart → stored file |
| `POST /assignments/{id}/analyze` | Analyze | 🟡 stub |
| `GET /assignments/{id}/results` | Results | 🟡 stub |
| `POST /feedback/{id}/resolve` | Resolve | 🟡 no-op |

---

## Product principles (§3)

| Principle | Status |
|-----------|--------|
| Student-controlled import | ✅ |
| Embedded at submit point | ✅ |
| Pre-submission, not auto-submit | ✅ |
| Visible content only | ✅ |
| Rubric-aware feedback | 🔴 display only, no AI |
| Explainable scoring | 🔴 Phase 7 |
| Minimal frontend complexity | ✅ |

---

## Recommended next up

1. **Discussion MVP (spec §5.6)** — prompt + main topic reply extraction on `/discussion_topics/`
2. **Phase 7** — one provider, one prompt, structured feedback UI
3. **Import metadata** — parse/store due date + points
4. **Phase 8** — re-analyze + resolve feedback
5. **Phase 9** — polish, settings, dashboard statuses
6. **Discussion v2** — thread reply composers + `draftEditorRole`

---

## How to update this checklist

When closing a gap:

1. Check the box or move 🟡 → ✅
2. Remove or shrink the matching **gap** bullet
3. Update **Summary** and **Last reviewed** date

When adding scope from spec §15 (future features), add under a separate **Post-MVP** section — do not mix into MVP phases unless promoted.
