/**
 * Shrink pruned fixture DOM to the minimum tests still need.
 */

const PORTAL_SELECTORS = [
  '[data-testid="assignment-rubric-modal"]',
  '[data-testid="enhanced-rubric-assessment-tray"]',
  '[role="menu"]',
  '[role="dialog"][aria-label="Criterion Long Description"]',
];

/** Drop Canvas layout wrappers; keep student content plus portaled menus/modals/trays. */
export function unwrapCanvasChrome(document) {
  const body = document.body;
  const studentView = body.querySelector('[data-testid="assignments-2-student-view"]');
  const topic = body.querySelector('[data-testid="discussion-topic-container"]');
  const root = studentView ?? topic;

  if (!root) {
    return;
  }

  const portaled = [];
  const seen = new Set();

  for (const selector of PORTAL_SELECTORS) {
    for (const el of body.querySelectorAll(selector)) {
      if (root.contains(el) || seen.has(el)) {
        continue;
      }
      if ([...seen].some((kept) => kept.contains(el))) {
        continue;
      }
      seen.add(el);
      portaled.push(portalOuterHtml(el));
    }
  }

  body.innerHTML = root.outerHTML + portaled.join("");
}

function portalOuterHtml(el) {
  if (el.matches('[data-testid="assignment-rubric-modal"]')) {
    return `<span data-testid="assignment-rubric-modal" role="dialog" aria-label="Assignment Rubric Details"><span class="css-eezqiu-closeButton"><button data-cid="BaseButton CloseButton" type="button"><span class="css-r9cwls-screenReaderContent">Close</span></button></span><button data-testid="preview-assignment-rubric-button" type="button">Preview Rubric</button></span>`;
  }

  if (el.matches('[data-testid="enhanced-rubric-assessment-tray"]')) {
    return el.outerHTML;
  }

  return el.outerHTML;
}

/** Fixture corpus needs this id; keep a tiny stub instead of full reply threads. */
export function ensureDiscussionRootEntryStub(document) {
  if (
    document.querySelector('[data-testid="discussion-topic-container"]') &&
    !document.querySelector('[data-testid="discussion-root-entry-container"]')
  ) {
    document.body.insertAdjacentHTML(
      "beforeend",
      '<div data-testid="discussion-root-entry-container"><div data-testid="discussion-entry-container">fixture replies</div></div>',
    );
  }
}

/** Replace long captured discussion prompts with a short stub that keeps extraction keywords. */
export function collapseDiscussionPrompts(document) {
  for (const el of document.querySelectorAll('[data-resource-type="discussion_topic.body"]')) {
    el.innerHTML =
      "<h2><strong>Arts Discussion Forum #9: The Role of Vulnerability, Risk, and Authenticity in Performance</strong></h2><p>Discussion prompt (fixture).</p>";
  }
}

/** Remove injected TinyMCE / RCE styles and toolbar chrome; keep textarea + iframe shell. */
export function collapseRichTextEditors(document) {
  for (const editor of document.querySelectorAll('[data-testid="text-editor"]')) {
    collapseAssignmentTextEditor(editor);
  }

  for (const composer of document.querySelectorAll('[data-testid="DiscussionEdit-container"]')) {
    collapseDiscussionComposer(composer);
  }
}

function collapseAssignmentTextEditor(editor) {
  const textarea = editor.querySelector("textarea");
  const textareaId = textarea?.id || "textentry_text";
  editor.innerHTML = `<textarea id="${textareaId}"></textarea><div class="tox-edit-area"><iframe class="tox-edit-area__iframe"></iframe></div>`;
}

function collapseDiscussionComposer(composer) {
  const textarea = composer.querySelector("textarea, [data-testid='message-body']");
  const textareaTag = textarea?.tagName.toLowerCase() === "textarea" ? "textarea" : "div";
  const textareaAttrs =
    textareaTag === "textarea"
      ? ` id="${textarea?.id || "message-body"}" data-testid="message-body"`
      : ` data-testid="message-body"`;
  const iframeHtml = composer.querySelector("iframe")
    ? `<div class="tox-edit-area"><iframe class="tox-edit-area__iframe"></iframe></div>`
    : "";

  const attachInput = composer.querySelector('[data-testid="attachment-input"]');
  const attachBtn = composer.querySelector('[data-testid="attach-btn"]');
  const removable = composer.querySelector('[data-testid="removable-item"]');
  const attachmentHtml = [
    attachInput ? `<input data-testid="attachment-input" type="file" />` : "",
    attachBtn?.outerHTML ?? "",
    removable ? minifyRemovableItem(removable) : "",
  ].join("");

  const cancel = composer.querySelector('[data-testid="DiscussionEdit-cancel"]');
  const submit = composer.querySelector('[data-testid="DiscussionEdit-submit"]');
  const actionsHtml = [cancel?.outerHTML, submit?.outerHTML].filter(Boolean).join("");

  composer.innerHTML = `<${textareaTag}${textareaAttrs}></${textareaTag}>${iframeHtml}${attachmentHtml}${actionsHtml}`;
}

function minifyRemovableItem(item) {
  const link = item.querySelector('a[href*="/files/"]');
  const href =
    link?.getAttribute("href") ?? "https://canvas.instructure.com/files/99543507/download?download_frd=1";
  const fromAria = item.getAttribute("aria-label")?.match(/^Replace\s+(.+?)\s+button$/i)?.[1];
  const fromHidden = link?.querySelector('[aria-hidden="true"] span[wrap="normal"]')?.textContent?.trim();
  const fileName = fromHidden || fromAria || "resume-1.pdf";
  return `<span role="button" aria-label="Replace ${fileName} button" data-testid="removable-item"><a href="${href}"><span aria-hidden="true"><span wrap="normal">${fileName}</span></span></a></span>`;
}

/** Keep one rubric criterion row and two rating bands — enough for scrape + points tests. */
export function minifyRubrics(document) {
  for (const table of document.querySelectorAll(
    '[data-testid="rubric-assessment-traditional-view"] table, [data-testid="rubric-preview"] table',
  )) {
    const tbody = table.querySelector("tbody");
    if (!tbody) {
      continue;
    }

    for (const row of [...tbody.querySelectorAll("tr")].slice(1)) {
      row.remove();
    }

    for (const ratings of tbody.querySelectorAll('[data-testid="traditional-view-criterion-ratings"]')) {
      for (const button of [...ratings.querySelectorAll("button[data-testid*='-ratings-']")].slice(2)) {
        button.remove();
      }
    }
  }
}

/** Submitted assignments embed a PDF preview iframe we do not need for DOM tests. */
export function stripSubmissionPreview(document) {
  for (const preview of document.querySelectorAll('[data-testid="assignments_2_submission_preview"]')) {
    preview.innerHTML = '<span class="css-r9cwls-screenReaderContent">fixture preview</span>';
  }
}

export function stripBodyStyleTags(document) {
  for (const style of document.body.querySelectorAll("style")) {
    style.remove();
  }
}

export function minifyFixtureDocument(document) {
  unwrapCanvasChrome(document);
  stripBodyStyleTags(document);
  collapseDiscussionPrompts(document);
  minifyRubrics(document);
  collapseRichTextEditors(document);
  stripSubmissionPreview(document);
  ensureDiscussionRootEntryStub(document);
}
