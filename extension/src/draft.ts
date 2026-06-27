import { findSubmitAnchor, isVisible } from "./injector";

type TinyMCEEditor = {
  getContent?: (options?: { format?: string }) => string;
  save?: () => void;
  id?: string;
};

type TinyMCEGlobal = {
  triggerSave?: () => void;
  editors?: TinyMCEEditor[];
  get?: (id: string) => TinyMCEEditor | null;
};

const EMPTY_EDITOR_HTML = "<p><br data-mce-bogus=\"1\"></p>";

function tinymceGlobal(): TinyMCEGlobal | undefined {
  return (window as Window & { tinymce?: TinyMCEGlobal }).tinymce;
}

function isEmptyEditorHtml(html: string): boolean {
  const trimmed = html.trim();
  return trimmed === "" || trimmed === EMPTY_EDITOR_HTML || trimmed === "<p></p>" || trimmed === "<p><br></p>";
}

function getSubmissionRoot(): Element {
  const submit = findSubmitAnchor();
  const fromSubmit = submit?.closest("#assignment_show, form.submit_assignment, .submission_form");
  if (fromSubmit) {
    return fromSubmit;
  }

  return document.querySelector("#assignment_show, .submission_form") ?? document.body;
}

function getSubmissionTextareas(root: Element): HTMLTextAreaElement[] {
  return Array.from(
    root.querySelectorAll<HTMLTextAreaElement>(
      "textarea#submission_body, textarea[name='submission[body]'], textarea[name*='submission'], textarea[id*='submission']",
    ),
  );
}

function syncSubmissionEditor(root: Element, textareas: HTMLTextAreaElement[]): void {
  const tinymce = tinymceGlobal();

  for (const textarea of textareas) {
    if (textarea.id && tinymce?.get) {
      tinymce.get(textarea.id)?.save?.();
    }
  }

  tinymce?.triggerSave?.();

  const active = document.activeElement;
  if (active instanceof HTMLElement && active !== document.body) {
    active.blur();
  }

  for (const textarea of textareas) {
    if (textarea.id && tinymce?.get) {
      tinymce.get(textarea.id)?.save?.();
    }
  }

  tinymce?.triggerSave?.();
}

function readEditorIframes(root: Element): string {
  const iframes = Array.from(
    root.querySelectorAll<HTMLIFrameElement>(
      ".tox-edit-area iframe, iframe.tox-edit-area__iframe, iframe[id*='tinymce']",
    ),
  ).filter((iframe) => isVisible(iframe));

  for (const iframe of iframes) {
    const body = iframe.contentDocument?.body;
    if (!body) {
      continue;
    }

    const html = body.innerHTML.trim();
    if (html && !isEmptyEditorHtml(html)) {
      return html;
    }
  }

  return "";
}

function readTinyMCEEditors(root: Element, textareas: HTMLTextAreaElement[]): string {
  const tinymce = tinymceGlobal();
  if (!tinymce) {
    return "";
  }

  for (const textarea of textareas) {
    if (!textarea.id || !tinymce.get) {
      continue;
    }

    const editor = tinymce.get(textarea.id);
    const html = editor?.getContent?.({ format: "html" })?.trim();
    if (html && !isEmptyEditorHtml(html)) {
      return html;
    }
  }

  for (const editor of tinymce.editors ?? []) {
    if (!editor.id) {
      continue;
    }

    const textarea = document.getElementById(editor.id);
    if (!textarea || !root.contains(textarea)) {
      continue;
    }

    const html = editor.getContent?.({ format: "html" })?.trim();
    if (html && !isEmptyEditorHtml(html)) {
      return html;
    }
  }

  return "";
}

function readSyncedTextarea(textareas: HTMLTextAreaElement[]): string {
  for (const textarea of textareas) {
    if (textarea.value.trim()) {
      return textarea.value.trim();
    }
  }

  return "";
}

/** Live editor read order (iframe → TinyMCE → textarea), not A2/classic page layout. */
export function extractDraftText(): string {
  const root = getSubmissionRoot();
  const textareas = getSubmissionTextareas(root);
  syncSubmissionEditor(root, textareas);

  // Live editor first — hidden textarea often still holds the last server draft.
  for (const reader of [
    () => readEditorIframes(root),
    () => readTinyMCEEditors(root, textareas),
    () => readSyncedTextarea(textareas),
  ]) {
    const content = reader();
    if (content) {
      return content;
    }
  }

  return "";
}
