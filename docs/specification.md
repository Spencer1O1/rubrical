# Rubrical Project Specification

> **Docs:** [README](../README.md) · [Development guide](./development.md) · [Spec checklist](./spec-checklist.md) (implementation status & gaps)

## 1. Project Vision

Rubrical is a pre-submission academic feedback tool that helps students evaluate their assignments against the actual assignment instructions and rubric before submitting. The goal is not to replace the student’s work, automate academic dishonesty, or submit assignments on the student’s behalf. Instead, Rubrical acts like a preflight checklist: it helps students see whether their draft satisfies the requirements, where it may lose points, and what specific improvements would make it stronger.

Many students lose points not because they are incapable of doing the work, but because they miss rubric details, misunderstand instructions, fail to include required elements, or submit work without a structured final review. Rubrical solves this by extracting the assignment context the student is already allowed to view, combining it with the student’s draft, and producing rubric-aware feedback.

The core experience should feel simple:

1. The student opens an assignment, discussion, or other supported Canvas page.
2. Rubrical injects a visible action button into the page, near the normal Canvas submission controls.
3. The student clicks “Check with Rubrical.”
4. Rubrical captures the visible assignment details, rubric, and current draft/submission text when available.
5. Rubrical analyzes the student’s work against the assignment instructions and rubric.
6. Rubrical gives structured feedback, estimated rubric performance, missing requirements, and revision suggestions.
7. The student revises and submits manually through Canvas.

Rubrical is not intended to be a Canvas API automation tool for v1. It should avoid needing Canvas API tokens, OAuth, or developer keys. Instead, it should use a browser extension to extract visible page content from the user’s active Canvas session only after the user intentionally clicks a Rubrical action.

## 2. Core Problem

Students often interact with assignments in scattered places:

* Assignment instructions are in Canvas.
* Rubrics may be in a separate table, modal, expandable panel, or lower section of the page.
* Required readings or prompts may be embedded in long text blocks.
* Drafts may be written in Google Docs, Word, VS Code, Canvas text editors, or discussion reply boxes.
* Submission requirements may be hidden in small details.
* Students often submit before comparing their work carefully against the rubric.

This creates a gap between “I wrote something” and “I know this satisfies the rubric.”

Rubrical closes that gap by turning unstructured assignment pages into structured grading context and comparing that context against the student’s current draft.

## 3. Product Principles

### 3.1 Student-Controlled

The student initiates every import and every analysis.

Rubrical may passively detect supported Canvas pages and inject a visible action button, but it must not import, transmit, store, or analyze assignment content, rubric content, or submission content until the user intentionally clicks a Rubrical action such as “Check with Rubrical” or “Import to Rubrical.”

### 3.2 Embedded Where the Student Already Works

Rubrical should not require the student to manually copy assignment instructions into a separate tool when the browser extension can safely read visible content from the current Canvas page.

The preferred UX is to place Rubrical directly next to the existing Canvas workflow:

```text
[Submit Assignment] [Check with Rubrical]
```

or:

```text
Before you submit:
[Check with Rubrical]
```

Rubrical should feel like a pre-submit checkpoint that appears at the moment the student needs it.

### 3.3 Pre-Submission, Not Auto-Submission

Rubrical helps the student review and improve their work. It should not submit assignments automatically in v1.

Manual submission keeps the student responsible for the final work and avoids unnecessary Canvas API/auth complexity.

### 3.4 Visible Content Only

The extension should only extract content visible to the authenticated user on the current Canvas page.

Rubrical should not:

* Read Canvas passwords.
* Read Canvas cookies.
* Extract session tokens.
* Use MITM techniques.
* Bypass Canvas permissions.
* Import hidden or unauthorized content.
* Scrape pages silently in the background.

### 3.5 Rubric-Aware Feedback

Feedback should be grounded in the assignment’s actual requirements and rubric, not generic writing advice.

The output should clearly connect suggestions to rubric criteria or assignment instructions.

### 3.6 Explainable Scoring

When Rubrical estimates performance, it should explain why.

Feedback should include:

* Evidence from the draft
* Missing elements
* Relevant rubric criteria
* Specific revision guidance
* Confidence level

### 3.7 Minimal Frontend Complexity

Rubrical should not become another large React/TypeScript application.

The browser extension should be tiny. The core application should be server-rendered using Go, templ, HTMX, and Tailwind.

JavaScript should only be used where the browser extension requires it.

## 4. Target Users

### 4.1 Primary User

A student who wants to check an assignment before submitting it.

Common needs:

* “Did I answer all parts of the prompt?”
* “Does my draft satisfy the rubric?”
* “What am I missing?”
* “What score would this probably receive?”
* “What should I revise before submitting?”
* “Did I include the required word count, citations, or examples?”
* “Is this ready to submit?”

### 4.2 Secondary User

A student-developer or power user who writes assignments outside Canvas and wants a better workflow for reviewing work before submission.

### 4.3 Future Users

Possible future users may include:

* Tutors
* Writing centers
* Teachers who want to preview rubric clarity
* Students working in groups
* Students using non-Canvas LMS platforms

These are not v1 priorities.

## 5. Core User Workflow

### 5.1 Canvas Page Detection Flow

1. User navigates to a supported Canvas page.
2. Browser extension content script detects page type.
3. Extension determines whether the page appears to be:

   * Assignment page
   * Discussion page
   * Assignment submission page
   * Discussion reply page
   * Rubric-bearing page
   * Unsupported page
4. If supported, the extension locates useful UI anchor points:

   * Submit Assignment button
   * Start Assignment button
   * Text entry editor
   * Discussion reply editor
   * Post Reply button
   * Rubric area
5. Extension injects a Rubrical button into the page.
6. No assignment, rubric, or draft content is transmitted yet.

The extension may inspect the page enough to decide where to place the button, but the import/analyze action must wait for the user click.

### 5.2 Preferred Button UX

Always use one label:

```text
Check with Rubrical
```

The button runs the same import flow whether or not Canvas exposes a readable draft (text editor, file selection, or assignment context only). Rubrical captures whatever is available and opens the review UI; missing draft content is handled there (paste or upload).

The button should be visually distinct from Canvas’s native submit action and should not trick the user into thinking it submits the assignment.

Suggested placement:

```text
[Submit Assignment] [Check with Rubrical]
```

or:

```text
Before you submit:
[Check with Rubrical]
```

### 5.3 Check with Rubrical Flow

1. User opens a Canvas assignment or discussion page.
2. Rubrical injects a “Check with Rubrical” button near the relevant submission controls.
3. User writes or previews their current draft in Canvas.
4. User clicks “Check with Rubrical.”
5. Extension extracts:

   * Page URL
   * Assignment/discussion title
   * Assignment instructions or discussion prompt
   * Visible rubric content
   * Due date if visible
   * Points possible if visible
   * Submission type if visible
   * Current draft/submission text if visible
6. Extension sends the extracted payload to the Rubrical backend.
7. Backend stores or updates an assignment snapshot.
8. Backend stores the current draft if one was detected.
9. Backend runs analysis if enough draft content is available.
10. Rubrical opens a results page or side panel with feedback.

### 5.4 No visible draft

If Canvas exposes no readable draft (file upload, external tool, empty editor), the **same** “Check with Rubrical” click still imports assignment/discussion context and rubric. Rubrical opens the review UI where the student pastes or uploads their draft, then runs analysis from there.

### 5.5 File Upload Assignment Flow

For file upload assignments, the Canvas page may not contain the student’s draft text.

Rubrical should handle this gracefully:

1. User opens the file upload assignment.
2. Extension injects “Check with Rubrical” near submit controls.
3. User clicks the button.
4. Extension imports assignment instructions and rubric.
5. Rubrical opens the review UI.
6. User uploads or pastes their draft into Rubrical.
7. Rubrical stores the submission file (when present) and analyzes it.

When the student has selected or uploaded a file on Canvas, **Check with Rubrical** reads that file from the Canvas page and stores it on the Rubrical server. The review UI shows the stored filename and upload time (not a blank “No file chosen” state). The student may replace the file or clear it from Rubrical without leaving the popup.

Future: read additional Canvas attachment types and LTI-hosted submissions where DOM access allows.

### 5.6 Discussion Flow

Discussions use the **same** Rubrical product flow as assignments (one label, one import, one review UI). They differ in Canvas layout and in what the student is often asked to write.

**Typical discussion requirements (often in one prompt):**

* **Topic response** — address the teacher’s prompt (substance, length, rubric criteria).
* **Participation rules** — e.g. “reply to at least 3 classmates with at least 25 words each.” These are still part of the imported prompt text; analysis must apply the right lens to the draft being checked.

**Editor-scoped capture (not a separate product mode):**

* Inject “Check with Rubrical” next to **reply composers the student is writing in**, not next to every classmate’s already-posted comment.
* Each click captures the **adjacent editor’s draft** plus shared context (title, full discussion prompt, rubric if present).
* Store drafts with `source_type` `canvas_discussion_reply` (and assignment snapshot `page_type` `discussion`).

**Planned draft context metadata** (extension → backend, for analysis):

* `draftEditorRole`: `topic_reply` | `thread_reply`
* Optional thread anchor (e.g. parent post id or excerpt) when replying to a classmate

**Phased scope:**

| Phase | Scope |
|-------|--------|
| **MVP (pre–Phase 7)** | `/discussion_topics/` — extract **discussion prompt** + **main topic reply** editor; verify end-to-end import like assignments. |
| **v2** | Additional buttons beside **thread reply** composers; set `draftEditorRole: thread_reply`. |
| **v2+ / Phase 7** | AI receives full prompt + `draftEditorRole` so peer replies are checked against participation rules, not the main-post rubric alone. Optional later: structured tagging of “participation requirements” vs “topic requirements” in stored context. |

Discussion snapshots reuse `assignment_snapshots` (same upsert, same review page). The checklist tracks discussion extraction separately from assignment extraction until MVP is verified on a real Canvas discussion.

### 5.7 Draft Analysis Flow

1. User opens an imported assignment in Rubrical.
2. User either:

   * Uses a draft captured from Canvas
   * Pastes a draft manually
   * Uploads a supported file in a future version
3. User clicks “Analyze” or the analysis starts automatically after the “Check with Rubrical” click if a draft was already captured.
4. Backend stores the draft as a `submission_draft`.
5. Backend normalizes the assignment instructions and rubric.
6. AI analysis service compares the draft to:

   * Assignment instructions or discussion prompt (full text, including participation rules when present)
   * Rubric criteria
   * Draft context when present (`draftEditorRole`: apply main-post vs peer-reply expectations appropriately)
   * Required word count
   * Required citations or sources
   * Required media/file/link elements, if detectable
7. Backend stores the analysis result.
8. HTMX updates the results panel with server-rendered feedback cards.

### 5.8 Revision Flow

1. User reviews feedback.
2. User edits their draft in Canvas, Rubrical, or their external writing tool.
3. User re-runs the analysis.
4. Rubrical shows the newest feedback and may show improvement from the previous analysis.
5. User manually submits final work in Canvas.

## 6. MVP Features

### 6.1 Browser Extension

The browser extension should be intentionally small.

Technology:

* TypeScript
* Manifest V3
* Content scripts
* Minimal popup if needed
* No React
* No heavy frontend framework

Responsibilities:

* Detect supported Canvas pages.
* Inject a Rubrical action button near useful Canvas controls.
* Extract visible assignment/discussion/rubric content after user click.
* Extract current visible draft text when available.
* Send extracted content to the Rubrical backend.
* Open the Rubrical results page or imported assignment page.

The extension should use a `MutationObserver` because Canvas may dynamically render page elements after initial load.

Initial supported page types:

* Canvas assignment pages
* Canvas online text entry assignments
* Canvas discussion pages
* Canvas discussion reply editors
* Canvas rubric tables when visible on the page

The extractor should use layered strategies:

1. Known Canvas selectors.
2. Semantic HTML regions such as `main`, `h1`, `table`, `article`, `textarea`, and contenteditable regions.
3. Text heuristics for labels such as “Rubric,” “Criteria,” “Pts,” “Due,” “Submitting,” “Submit Assignment,” and “Reply.”
4. Manual fallback in a future version.

### 6.2 Button Injection

Rubrical injects one visible button on supported Canvas pages:

```text
Check with Rubrical
```

Injection targets:

* Near Submit Assignment / Post Reply / discussion reply controls
* Near the **active reply composer** being checked (see §5.6 — one button per editor, not per existing thread comment)
* Near online text entry editor
* Near rubric area if no submission controls are found

Rules:

* Do not inject duplicate buttons.
* Do not hide or replace Canvas controls.
* Do not make Rubrical’s button look like the official submit button.
* Do not submit the assignment.
* Do not transmit page content until the user clicks the button.

### 6.3 Assignment Snapshot Storage

Rubrical should store imported assignment snapshots so the user can return to them later.

Stored data should include:

* Source URL
* Source platform
* Page type
* Imported title
* Raw visible text
* Extracted instructions
* Extracted rubric rows
* Due date if found
* Points possible if found
* Submission type if found
* Import timestamp

### 6.4 Current Draft Extraction

When available, the extension should extract the current student draft from the Canvas page after the user clicks “Check with Rubrical.”

Possible draft sources:

* Textarea
* Rich text editor iframe
* Contenteditable editor
* Discussion reply box
* Assignment text entry editor
* Visible submitted text preview

If no draft is detected, Rubrical should still import the assignment context and prompt the user to paste or upload their draft.

### 6.5 Draft Input Inside Rubrical

The user should be able to paste a draft into Rubrical if the extension cannot extract one.

**Stored submission files (required):**

Rubrical always persists the student’s submission **file bytes** when one is captured from Canvas or uploaded in the UI. Metadata lives on `submission_drafts`; bytes live in local object storage (`RUBRICAL_DATA_DIR`, default `./data`).

* Original filename (`source_file_name`)
* MIME type, byte size, upload timestamp
* Storage key pointing at the on-disk blob

**Pasted / Canvas text (`body`):**

`body` holds **only** text captured from a Canvas editor or pasted/edited in the Rubrical textarea. Rubrical does **not** extract text from uploaded files into `body`. When a submission file is stored, Phase 7 analysis receives the **file** — including `.txt`, `.md`, `.docx`, and `.pdf`. Models handle long text better from the original file than from a lossy extraction step.

**UI:**

Rubrical mirrors Canvas: **one active submission type at a time** — Text, File, or Web URL. Switching types clears the other type’s stored content (same as Canvas).

* Segmented control (tab-ish) at the top of the draft panel: **Text | File | Web URL** — **only tabs the assignment allows** (from Canvas `submission_types` / submission-type selector DOM, stored on the snapshot)
* If only one type is allowed, hide the switcher and show that panel directly
* If allowed types were not captured on import, default to all three tabs (text, file, URL)
* On import, default to the Canvas tab that is selected (from `draftKind` / captured content)
* **Web URL capture:** Canvas A2 uses `input[data-testid="url-input"]`
* **Text:** textarea + Analyze
* **File:** stored filename + upload time, Remove, Replace via file picker
* **Web URL:** URL input + Save

`body`, stored file bytes, and `submission_url` are **mutually exclusive** per draft row (`draft_mode`).

Future draft input:

* Google Docs import
* Canvas submission box extraction improvements

### 6.5.1 Canvas file staging (extension)

For file-upload assignments, the extension stages submission file **bytes in Canvas page-origin IndexedDB** (content script context) when the student selects files on Canvas or removes them from the upload table. Staging is cleared after a successful **Check with Rubrical** click.

Large files (up to **500 MB** per file, aligned with server draft upload limits) are written directly as `Blob` values — they are **not** sent through the extension service worker, which avoids Chrome message-size limits.

**Staging keys:**

* **Primary:** Canvas file id (numeric trash-button `id`) once the upload table row exists.
* **Provisional:** normalized filename + `stagedAt` (ISO timestamp at file-input capture) until the Canvas id is known; then promote provisional → id.

**Page load (metadata only — required for row indicators):**

On assignment page boot, after assignment context is ready, the extension may **`GET /assignments/draft-manifest?sourceUrl=…`** for the normalized current page URL. Response includes filenames, `serverFileId`, `byteSize`, and `uploadedAt` only — **no file bytes**. Empty when the assignment was never imported.

The extension merges manifest metadata with IndexedDB staging and Canvas `uploaded_files_table` rows. Rows Rubrical **cannot** read (no staged bytes and no manifest match) show a **re-upload** indicator (×, tooltip “Re-upload to use in Rubrical”). Rows that are staged locally or already saved on Rubrical show **no** indicator.

**Click import:**

`POST /imports` writes assignment context and draft metadata. For **assignment file uploads**, staged bytes are **not** inlined in the JSON body — after import succeeds, the extension uploads each staged file via **`POST /assignments/{id}/draft/upload`** (multipart) from the content script. **`draftFileRefs`** (existing server file ids for rows still on Canvas) may still appear in the import JSON. **No Canvas file API download** — bytes come only from extension staging or server refs.

Discussion attachments may still carry inline base64 in `draftFiles` when small enough; large discussion attachments follow the same multipart path after import.

**Hooks:** file input `change` (stage bytes); trash-button click (delete staged row); canvas file id assignment (reconcile provisional → id when trash `button[id]` appears). The extension does **not** read Canvas upload-progress UI — bytes from the file picker are sufficient until click.

**Indicators:** re-upload × only when a row has no staged bytes and no server manifest match.

**Multi-device:** server manifest enables reuse via `draftFileRefs` on another device after a prior click; IndexedDB staging is per-browser and per-Canvas-origin only (clearing Canvas site data clears staged bytes).

**Service worker role:** the extension service worker proxies Rubrical API `fetch` / multipart requests from Canvas pages (Private Network Access). It does **not** store staged submission files.

### 6.6 Rubric Normalization

The backend should attempt to transform rubric tables into structured criteria.

A criterion may include:

* Criterion name
* Description
* Point value
* Rating levels
* Rating descriptions
* Raw extracted text

Not every Canvas rubric will be clean. The system should preserve raw extracted data even when normalization is imperfect.

### 6.7 AI Analysis

The AI layer should produce structured results, not just a blob of prose.

Analysis output should include:

* Overall summary
* Estimated rubric performance
* Criteria-by-criteria feedback
* Missing requirements
* Strengths
* Specific revision suggestions
* Risk flags
* Word count verification when applicable
* Evidence from the student draft
* Confidence level

The AI should be instructed to avoid writing the assignment for the student. It should focus on evaluation, diagnosis, and revision guidance.

### 6.8 Feedback UI

The results page should show feedback in clear sections:

* Overall readiness
* Estimated score or rubric level
* Missing requirements
* Rubric criteria cards
* Suggested revisions
* Strengths
* Warnings

Each feedback card should be easy to scan.

Example feedback card:

* Criterion: Uses assigned reading
* Status: Partially met
* Estimated score: 3/5
* Evidence found: Mentions Small’s idea of “musicking”
* Missing: Does not connect Pearson’s idea of the ephemeral to the live event example
* Suggested revision: Add 2–3 sentences connecting silence, breath, and audience presence to Pearson’s argument

### 6.9 Dashboard

The dashboard should show the user’s imported assignments.

Each item should display:

* Assignment title
* Course/source if available
* Import date
* Last analysis date
* Due date if available
* Status:

  * Imported
  * Draft captured
  * Draft added
  * Analyzed
  * Revised
  * Ready

## 7. Non-Goals for v1

Rubrical v1 should not:

* Automatically submit assignments.
* Extract Canvas cookies or session tokens.
* Use MITM techniques.
* Require Canvas API tokens.
* Require Canvas OAuth.
* Impersonate the user outside the visible browser page.
* Scrape content in the background without user action.
* Import or analyze pages without a user click.
* Guarantee a grade.
* Write full assignments for the student.
* Support every LMS.
* Build a complex SPA frontend.
* Include collaborative editing.
* Include teacher/admin dashboards.

## 8. Technical Stack

### 8.1 Browser Extension

Technology:

* TypeScript
* Manifest V3
* Minimal popup UI
* Content scripts
* MutationObserver-based page detection
* No React
* No heavy frontend framework

Responsibilities:

* Page detection
* Button injection
* DOM extraction after user click
* Draft detection
* Data packaging
* POST to Rubrical backend
* Open Rubrical assignment/results page

The extension should remain small and replaceable. All complex logic belongs in the Go backend.

### 8.2 Backend

Technology:

* Go
* chi router
* Standard `net/http` patterns where practical

Responsibilities:

* HTTP routing
* Assignment imports
* Draft storage
* Rubric normalization
* AI orchestration
* Database access
* Server-rendered page responses
* HTMX partial responses
* Authentication if added later

Suggested backend structure:

```text
rubrical/
  cmd/
    server/
      main.go

  internal/
    web/
      routes.go
      handlers/
      pages/
      components/

    db/
      queries/
      migrations/

    canvasextract/
      normalize.go
      selectors.go
      rubric.go
      draft.go

    analysis/
      analyzer.go
      prompts.go
      schemas.go

    ai/
      client.go
      openai.go
      anthropic.go

    auth/
      session.go

    config/
      config.go

  extension/
    manifest.json
    src/
      content.ts
      popup.ts
      injector.ts
      extractor.ts
      draft.ts

  migrations/
  sql/
  static/
  templates-or-templ/
```

### 8.3 Routing

Use `chi` for routing.

Example routes:

```text
GET  /                         Dashboard
POST /imports                  Import assignment from extension
GET  /assignments/draft-manifest  Draft file metadata for extension (query: sourceUrl)
GET  /assignments/{id}         Assignment detail page
POST /assignments/{id}/draft   Save draft
POST /assignments/{id}/analyze Analyze draft
GET  /assignments/{id}/results Results panel
POST /feedback/{id}/resolve    Mark feedback item resolved
GET  /health                   Health check
```

HTMX routes should return HTML fragments when appropriate.

Example:

```text
POST /assignments/{id}/analyze
```

This route should:

1. Save the submitted draft.
2. Run analysis.
3. Store analysis results.
4. Return the updated feedback panel as an HTML fragment.

### 8.4 HTML Rendering

Technology:

* templ

Rubrical should use templ for type-safe server-rendered HTML components.

Component examples:

* `Layout`
* `DashboardPage`
* `AssignmentPage`
* `DraftEditor`
* `AnalysisPanel`
* `FeedbackCard`
* `RubricCriterionCard`
* `ImportStatusBadge`
* `ScoreEstimateBadge`

The UI should be built from composable server-rendered components.

### 8.5 Interactivity

Technology:

* HTMX

HTMX should handle:

* Submitting drafts
* Running analysis
* Updating feedback panels
* Marking feedback items as resolved
* Refreshing assignment status
* Loading partial content
* Showing progress/loading states

The project should avoid building unnecessary client-side state management.

Example UI behavior:

```html
<form
  hx-post="/assignments/123/analyze"
  hx-target="#analysis-results"
  hx-swap="innerHTML"
>
  <textarea name="draft"></textarea>
  <button type="submit">Analyze</button>
</form>

<section id="analysis-results">
  <!-- Server-rendered feedback appears here -->
</section>
```

### 8.6 Database

Technology:

* PostgreSQL

Rubrical should use Postgres as the primary database.

Core tables:

* users
* assignment_snapshots
* rubric_criteria
* submission_drafts
* analysis_runs
* analysis_attempts
* feedback_items
* extracted_sources

For local development, the app can run against a local Postgres container.

### 8.7 Migrations

Technology:

* goose

Migrations should live in the `migrations/` directory and be committed to the repository.

Migration naming (solo dev — one squashed file in git):

```text
00001_initial_schema.sql
```

While iterating locally you may add temporary `00002_*.sql` files; before finishing, fold everything into `00001_initial_schema.sql`, mirror the same shape in `sql/schema/schema.sql`, delete incrementals, and run `make db-reset`. See `docs/development.md`.

### 8.8 Queries

Technology:

* sqlc

SQL should be written explicitly and compiled into type-safe Go code using sqlc.

Suggested layout:

```text
sql/
  queries/
    assignments.sql
    drafts.sql
    analysis.sql
    feedback.sql
  schema/
```

sqlc should generate Go database access code into an internal package such as:

```text
internal/db/gen
```

### 8.9 Styling

Technology:

* Tailwind CSS

Tailwind should be used for styling server-rendered templ components.

The UI should feel clean, focused, and academic without becoming visually heavy.

Design direction:

* Minimal dashboard
* Strong typography for assignment text
* Clear feedback cards
* Status badges
* Rubric-aligned color/priority indicators
* Comfortable spacing for long-form reading
* Mobile-friendly, but desktop-first for v1

### 8.10 AI Service Layer

Technology:

* Go service layer
* Provider adapters for OpenAI, Anthropic, or other models

The AI layer should be abstracted behind an internal interface.

Example:

```go
type Analyzer interface {
    AnalyzeDraft(ctx context.Context, input AnalysisInput) (AnalysisResult, error)
}
```

Provider-specific code should not leak into handlers.

The analysis service should:

* Build prompts from structured assignment data.
* Include rubric criteria.
* Include the student draft.
* Request structured JSON output.
* Validate the model response.
* Store the result.
* Return renderable feedback data to the web layer.

The system should support multiple providers later.

Possible providers:

* OpenAI
* Anthropic
* Local model later, if practical

## 9. Suggested Data Model

### 9.1 users

Stores application users if auth is added.

Fields:

* id
* email
* display_name
* created_at
* updated_at

For v1 local-only development, user accounts may be skipped or replaced with a single local user.

### 9.2 assignment_snapshots

Stores imported Canvas assignment data.

Fields:

* id
* user_id
* source_url
* source_platform
* page_type
* course_name
* assignment_title
* raw_text
* instructions_text
* due_at
* points_possible
* submission_type
* imported_at
* created_at
* updated_at

### 9.3 rubric_criteria

Stores normalized rubric criteria.

Fields:

* id
* assignment_snapshot_id
* name
* description
* points_possible
* ratings_json
* raw_text
* sort_order
* created_at
* updated_at

### 9.4 submission_drafts

Stores student draft **text/URL mode** metadata. File bytes live in `submission_draft_files` (below).

Fields:

* id
* assignment_snapshot_id
* user_id
* body — Canvas-captured or manually pasted text only (not extracted from files)
* word_count
* source_type
* draft_mode — `text`, `file`, or `url` (mutually exclusive active submission)
* submission_url — website URL when `draft_mode = url`
* captured_from_canvas
* created_at
* updated_at

### 9.4.1 submission_draft_files

Stores submission file metadata for multi-file drafts. One draft row may have many file rows.

Fields:

* id
* submission_draft_id
* source_file_name
* file_storage_key — path relative to `RUBRICAL_DATA_DIR`
* file_mime_type
* file_byte_size
* uploaded_at
* sort_order
* created_at

File bytes are **not** in Postgres; they live under `RUBRICAL_DATA_DIR` at `file_storage_key`.

Possible `source_type` values:

* canvas_text_entry
* canvas_file_upload
* canvas_website_url
* canvas_discussion_reply
* manual_paste
* file_upload
* imported_submission_preview

### 9.5 analysis_runs

Stores the **current** completed analysis for an assignment. When a new analyze succeeds, previous runs for that assignment are deleted (feedback cascades).

Fields:

* id
* assignment_snapshot_id
* submission_draft_id
* provider
* model
* status
* overall_summary
* predicted_score
* predicted_score_max
* confidence
* raw_model_input
* raw_model_output
* created_at
* completed_at

### 9.5a analysis_attempts

Append-only log of every analyze request for rate limiting. Not deleted when a new analysis replaces the current one.

Fields:

* id
* user_id
* assignment_snapshot_id
* analysis_run_id
* status (`started`, `completed`, `failed`)
* created_at
* completed_at

### 9.6 feedback_items

Stores individual feedback items from analysis.

Fields:

* id
* analysis_run_id
* rubric_criterion_id
* category
* severity
* title
* explanation
* score_rationale
* fulfilled_requirements
* unfulfilled_requirements
* criterion_status
* criterion_score
* predicted_points
* max_points
* selected_rating
* status
* sort_order
* created_at
* updated_at

Feedback categories may include:

* strength
* guidance

Severity values may include:

* info
* low
* medium
* high
* critical

### 9.7 extracted_sources

Stores raw and normalized extraction records.

Fields:

* id
* assignment_snapshot_id
* source_kind
* raw_content
* normalized_content
* extraction_method
* confidence
* created_at

Possible `source_kind` values:

* assignment_instructions
* discussion_prompt
* rubric_table
* due_date
* points_possible
* submission_type
* visible_draft

## 10. Extension Extraction Model

The extension should output a structured payload.

Example import payload:

```json
{
  "sourceUrl": "https://school.instructure.com/courses/123/assignments/456",
  "pageType": "assignment",
  "title": "Arts Discussion Forum #7",
  "visibleText": "...",
  "instructionsText": "...",
  "draftText": "...",
  "draftKind": "file",
  "draftFiles": [
    {
      "fileName": "essay.pdf",
      "mimeType": "application/pdf",
      "contentBase64": "..."
    }
  ],
  "draftFileRefs": [
    {
      "serverFileId": 7,
      "fileName": "notes.txt",
      "canvasFileId": "99543121",
      "sortOrder": 0
    }
  ],
  "draftEditorRole": "topic_reply",
  "rubricRows": [
    ["Criteria", "Ratings", "Pts"],
    ["Reflection depth", "Full credit...", "10 pts"],
    ["Word count", "Minimum 300 words", "5 pts"]
  ],
  "metadata": {
    "dueDateText": "Due Jun 26",
    "pointsPossibleText": "25 pts",
    "submissionTypeText": "Online Text Entry"
  },
  "capturedAt": "2026-06-26T12:00:00Z"
}
```

The backend should not trust the extractor blindly. It must validate and normalize every import:

* `sourceUrl` — normalized, `http(s)` Canvas `instructure.com` assignment or discussion URL
* Size limits on text fields, rubric rows/ratings, and base64 draft files
* Allowed `pageType` values: `assignment`, `discussion`, `unknown`
* Rubric structure bounds; invalid base64 rejected
* `dueDateText` / `pointsPossibleText` parsed best-effort into `due_at` / `points_possible` (null when parse fails)
* Instructions HTML sanitized before storage (existing behavior)

`draftEditorRole` is optional until thread-reply capture ships (§5.6). Values: `topic_reply`, `thread_reply`. Omitted or `topic_reply` for MVP discussion work.

### 10.1 Draft Detection Rules

The extension should attempt to detect drafts from:

* `textarea`
* `contenteditable="true"`
* rich text editor body
* visible discussion reply editor
* visible assignment text entry editor
* visible previous submission content

If multiple candidates are found, the extension should choose the most likely active editor and may include metadata about extraction confidence.

## 11. AI Analysis Output Schema

The AI returns structured JSON per criterion: `selectedRating`, `bandPosition` (0–100 within that band), `scoreRationale`, `fulfilledRequirements`, and `unfulfilledRequirements` (one suggestion per gap). Top-level `strengths` and `guidance` summarize cross-cutting feedback. The server maps band + bandPosition to a continuous 0–1 position for the gradient arrow, assigns that band's points, derives `status`, and sums `predictedScore`.

Example conceptual output:

```json
{
  "overallSummary": "The draft addresses the main prompt and includes a specific live performance example, but it could more directly connect Pearson's idea of ephemerality to the described performance.",
  "confidence": "medium",
  "criteria": [
    {
      "criterionName": "Connects to course concepts",
      "selectedRating": "Full Credit",
      "bandPosition": 18,
      "scoreRationale": "The draft reaches the full-credit band because it directly discusses Small's concept of musicking, but it sits low in the band because the Pearson connection is only implied.",
      "fulfilledRequirements": [
        {
          "requirement": "Discusses Small's idea of musicking",
          "evidence": "The draft identifies performers, guests, relatives, and staff as contributors to the musical event."
        }
      ],
      "unfulfilledRequirements": [
        {
          "requirement": "Connects Pearson's idea of ephemerality to the live event",
          "severity": "medium",
          "explanation": "The draft describes the atmosphere and audience laughter, but does not explicitly tie those details to Pearson's idea that performance exists only in the moment.",
          "suggestion": "Add a direct sentence explaining how the pauses, laughter, and shared attention made the performance unrepeatable."
        }
      ]
    }
  ],
  "strengths": [
    "The draft uses a specific personal live-performance example."
  ],
  "guidance": [
    "Make the Pearson connection explicit because it is the highest-impact missing course concept."
  ]
}
```

The backend should parse this into database records and render it through templ components.

## 12. Security and Privacy

Rubrical may handle sensitive academic data. The app should be designed with privacy in mind from the beginning.

Rules:

* Do not collect Canvas passwords.
* Do not collect Canvas access tokens.
* Do not collect session cookies.
* Do not use MITM techniques.
* Do not auto-submit assignments.
* Do not import pages without user action.
* Do not analyze drafts without user action.
* Do not expose drafts publicly.
* Do not log full drafts unnecessarily.
* Keep AI API keys server-side only.
* Use environment variables for secrets.
* Redact sensitive content from application logs.
* Add clear user messaging about what is being imported and analyzed.

For v1, Rubrical should assume that imported assignment text and student drafts are private user data.

### 12.1 Passive Detection vs Active Import

Rubrical may passively detect Canvas page structure to inject a button.

Rubrical must not passively transmit or store assignment/draft content.

Allowed passive behavior:

* Detecting page type
* Looking for submit buttons
* Looking for editors
* Looking for rubric areas
* Injecting a button
* Updating button text based on context
* **`GET /assignments/draft-manifest`** for the normalized current page URL — **metadata only** (filenames, ids, sizes); empty when never imported

Not allowed passively:

* `POST /imports` or any draft/file **byte** upload
* Sending assignment text to backend
* Sending draft text to backend
* Sending rubric text to backend
* Running AI analysis
* Storing a page snapshot

Active import (including submission file bytes) begins only after the user clicks **Check with Rubrical**.

## 13. Development Environment

Suggested development setup:

```text
Go server running locally
Postgres running in Docker
Tailwind build/watch process
templ generate/watch process
Browser extension loaded unpacked in Chrome
```

Example services:

```text
localhost:8787       Rubrical web app
localhost:5432       Postgres
Chrome extension     Imports Canvas page content
```

Suggested Makefile commands:

```text
make dev             Run server, templ, and Tailwind watchers
make db-up           Start Postgres
make migrate-up      Run goose migrations
make sqlc            Generate sqlc code
make test            Run Go tests
make extension-build Build extension assets
```

## 14. MVP Build Order

### Phase 1: Skeleton

* Create Go project.
* Add chi router.
* Add templ.
* Add Tailwind.
* Add basic dashboard page.
* Add Postgres Docker setup.
* Add goose.
* Add sqlc.

### Phase 2: Browser Extension Injection

* Create Manifest V3 browser extension.
* Add content script.
* Detect Canvas assignment/discussion pages.
* Use MutationObserver to handle dynamic page rendering.
* Inject “Check with Rubrical” button.
* Prevent duplicate injection.
* Confirm that no content is transmitted before click.

### Phase 3: Assignment Import

* On button click, extract page title and main visible text.
* Extract assignment instructions or **discussion prompt** (MVP: main topic on `/discussion_topics/`).
* Extract due date and points if visible.
* POST import payload to backend.
* Store assignment snapshot (`page_type` `assignment` or `discussion`).
* Show imported assignment page.

### Phase 4: Draft Capture

* Detect text entry editors and **discussion topic reply** editor (MVP).
* Extract current draft after button click from the editor adjacent to the clicked button.
* Store captured draft (`source_type` as appropriate).
* If no draft is detected, show paste/upload prompt in Rubrical.
* **Later (v2):** thread reply composers + `draftEditorRole: thread_reply`.

### Phase 5: Rubric Extraction

* Extract table rows from Canvas.
* Store raw rubric rows.
* Normalize into rubric criteria.
* Display rubric criteria in Rubrical.

### Phase 6: Draft Input

* Add draft textarea in Rubrical.
* Save draft.
* Count words.
* Display draft status.

### Phase 7: AI Analysis

* Create AI provider interface.
* Add OpenAI or Anthropic adapter.
* Build first analysis prompt.
* **Submission input:** when `file_storage_key` is set, pass the stored file. When `draft_mode = url`, pass `submission_url` (Phase 7 may fetch page content). When `draft_mode = text`, pass `body`.
* For discussions: pass full prompt + `draftEditorRole` so peer replies use participation expectations, not main-post rubric alone.
* Request structured JSON.
* Store analysis run.
* Render feedback cards.

### Phase 8: Revision Loop

* Allow re-analysis.
* Show analysis history.
* Mark feedback items resolved.
* Compare previous and latest results.

### Phase 9: Polish

* Improve Canvas selectors.
* Add manual extraction fallback.
* Improve UI.
* Add HTMX loading states.
* Add error handling.
* Add privacy/settings page.

## 15. Future Features

Possible post-MVP features:

* Manual region selection from Canvas pages
* PDF instruction extraction
* Google Docs import
* Assignment requirement checklist
* Structured participation-requirement tagging (discussion prompts)
* Essay mode
* Lab report mode
* Code assignment mode
* Citation checker
* Source requirement checker
* Local-only desktop mode
* Native messaging bridge
* Support for other LMS platforms
* Teacher-facing rubric clarity checker
* Team/group project review mode
* Multiple AI provider selection
* Bring-your-own API key mode
* Optional Canvas API/OAuth integration if institutionally available
* Optional side panel mode instead of opening a new tab

## 16. Brand and Positioning

Rubrical should be positioned as:

> A rubric-aware preflight checker for student assignments.

Possible tagline:

> Check the rubric before the rubric checks you.

Alternative positioning:

> Rubrical helps students review their work against the actual assignment instructions before submitting.

The tone should be helpful, direct, and student-centered. It should avoid sounding like a cheating tool. The product should emphasize learning, revision, and responsibility.

## 17. Summary

Rubrical is a Canvas-adjacent academic review tool that helps students check their work before submission. It injects a visible “Check with Rubrical” action into supported Canvas assignment and discussion pages, then imports visible assignment, rubric, and draft content only after the student intentionally clicks the button.

The technical architecture intentionally avoids frontend complexity:

* TypeScript browser extension, no React
* Go backend with chi
* Server-rendered HTML with templ
* HTMX for interactivity
* Postgres for persistence
* goose for migrations
* sqlc for type-safe queries
* Tailwind for styling
* Go AI service layer for OpenAI/Anthropic integration

The central philosophy is simple: Rubrical should help students become more aware of requirements before they submit. It should not bypass Canvas, steal credentials, auto-submit work, or replace the student’s authorship. It should be a reliable academic preflight tool embedded at the point where students need it most.
