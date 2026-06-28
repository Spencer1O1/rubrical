import { discussion, submission } from "./canvas/anchors";
import type { CanvasAnchor } from "./canvas/anchors/types";
import { firstMatch, queryAnchor, queryAnchorAll } from "./canvas/query";
import { findSubmitAnchor, isVisible } from "./injector";
import { isStrictExtraction } from "./strict";

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

type DraftEditorAnchors = {
  root: Element;
  textareas: CanvasAnchor;
  editorIframes: CanvasAnchor;
};

const EMPTY_EDITOR_HTML = "<p><br data-mce-bogus=\"1\"></p>";

function tinymceGlobal(): TinyMCEGlobal | undefined {
  return (window as Window & { tinymce?: TinyMCEGlobal }).tinymce;
}

function isEmptyEditorHtml(html: string): boolean {
  const trimmed = html.trim();
  return trimmed === "" || trimmed === EMPTY_EDITOR_HTML || trimmed === "<p></p>" || trimmed === "<p><br></p>";
}

function discussionEditRoot(): Element | null {
  return queryAnchor(discussion.editContainer);
}

function resolveDraftEditorAnchors(): DraftEditorAnchors {
  const discussionRoot = discussionEditRoot();
  if (discussionRoot) {
    return {
      root: discussionRoot,
      textareas: discussion.textareas,
      editorIframes: discussion.editorIframes,
    };
  }

  const assignmentRoot = queryAnchor(submission.root) ?? document.body;
  return {
    root: queryAnchor(submission.textEditor, assignmentRoot) ?? assignmentRoot,
    textareas: submission.textareas,
    editorIframes: submission.editorIframes,
  };
}

export function getSubmissionRoot(): Element {
  const discussionRoot = discussionEditRoot();
  if (discussionRoot) {
    return discussionRoot;
  }

  const anchored = queryAnchor(submission.root);
  if (anchored) {
    return anchored;
  }

  if (!isStrictExtraction()) {
    const submit = findSubmitAnchor();
    const fromSubmit = submit?.closest([...submission.root.classic].join(", "));
    if (fromSubmit) {
      return fromSubmit;
    }

    const classic = firstMatch(submission.root.classic);
    if (classic) {
      return classic;
    }
  }

  return document.body;
}

function getSubmissionTextareas(anchors: DraftEditorAnchors): HTMLTextAreaElement[] {
  return queryAnchorAll<HTMLTextAreaElement>(anchors.textareas, anchors.root);
}

function syncSubmissionEditor(textareas: HTMLTextAreaElement[]): void {
  const tinymce = tinymceGlobal();
  if (!tinymce) {
    return;
  }

  const saveById = (id: string) => {
    tinymce.get?.(id)?.save?.();
  };

  for (const textarea of textareas) {
    if (textarea.id) {
      saveById(textarea.id);
    }
  }

  for (const editor of tinymce.editors ?? []) {
    if (editor.id) {
      saveById(editor.id);
    }
  }

  tinymce.triggerSave?.();
}

function readEditorIframes(anchors: DraftEditorAnchors): string {
  const iframes = queryAnchorAll<HTMLIFrameElement>(anchors.editorIframes, anchors.root).filter(
    (iframe) => isVisible(iframe),
  );

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

function readTinyMCEEditors(anchors: DraftEditorAnchors, textareas: HTMLTextAreaElement[]): string {
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
    if (!textarea || !anchors.root.contains(textarea)) {
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
  const anchors = resolveDraftEditorAnchors();
  const textareas = getSubmissionTextareas(anchors);
  syncSubmissionEditor(textareas);

  for (const reader of [
    () => readEditorIframes(anchors),
    () => readTinyMCEEditors(anchors, textareas),
    () => readSyncedTextarea(textareas),
  ]) {
    const content = reader();
    if (content) {
      return content;
    }
  }

  return "";
}
