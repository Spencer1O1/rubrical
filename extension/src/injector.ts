import { isSupportedCanvasPath, submitAnchorOrder } from "./canvas/anchors";
import { queryAnchor } from "./canvas/query";

export const BUTTON_ID = "rubrical-action-button";

let lastMirroredSignature = "";
let lastLabel = "";

function childElements(el: Element): Element[] {
  return Array.from(el.children);
}

function elementAttributes(el: Element): Attr[] {
  return Array.from(el.attributes);
}

export function isSupportedCanvasPage(): boolean {
  if (!/instructure\.com/.test(window.location.hostname)) {
    return false;
  }

  const path = window.location.pathname;
  return isSupportedCanvasPath(window.location.pathname);
}

export function detectPageType(): string {
  const path = window.location.pathname;
  if (path.includes("/assignments/")) return "assignment";
  if (path.includes("/discussion_topics/")) return "discussion";
  return "unknown";
}

export function isVisible(el: HTMLElement): boolean {
  const style = window.getComputedStyle(el);
  if (style.display === "none" || style.visibility === "hidden") {
    return false;
  }
  const rect = el.getBoundingClientRect();
  return rect.width > 0 && rect.height > 0;
}

export function findSubmitAnchor(): HTMLButtonElement | null {
  for (const anchor of submitAnchorOrder) {
    const el = queryAnchor<HTMLButtonElement>(anchor);
    if (el && isVisible(el)) {
      return el;
    }
  }
  return null;
}

export function needsPlacement(): boolean {
  const submit = findSubmitAnchor();
  if (!submit) {
    return false;
  }

  const button = document.getElementById(BUTTON_ID);
  if (!button) {
    return true;
  }

  return button.nextElementSibling !== submit;
}

function structureSignature(el: Element): string {
  const parts: string[] = [];
  const walk = (node: Element): void => {
    parts.push(`${node.tagName.toLowerCase()}:${node.className}`);
    for (const child of childElements(node)) {
      walk(child);
    }
  };
  walk(el);
  return parts.join(">");
}

/** Recursively clone submit's inner element tree; classes come from Canvas, not hardcoded. */
function cloneElementStructure(source: Element, label: string): Element {
  const clone = document.createElement(source.tagName.toLowerCase());
  clone.className = source.className;

  for (const { name, value } of elementAttributes(source)) {
    if (name === "class" || name === "id") {
      continue;
    }
    clone.setAttribute(name, value);
  }

  const childElementsList = childElements(source);
  if (childElementsList.length === 0) {
    clone.textContent = label;
    return clone;
  }

  for (const child of childElementsList) {
    clone.appendChild(cloneElementStructure(child, label));
  }

  return clone;
}

function cloneSubmitContent(submit: HTMLButtonElement, label: string): DocumentFragment {
  const fragment = document.createDocumentFragment();
  for (const child of childElements(submit)) {
    fragment.appendChild(cloneElementStructure(child, label));
  }
  return fragment;
}

function findLabelElement(root: Element): HTMLElement | null {
  let deepest: Element = root;
  let maxDepth = -1;

  const walk = (el: Element, depth: number): void => {
    if (depth > maxDepth) {
      maxDepth = depth;
      deepest = el;
    }
    for (const child of childElements(el)) {
      walk(child, depth + 1);
    }
  };

  walk(root, 0);
  return deepest instanceof HTMLElement ? deepest : null;
}

/** Mirror Canvas button chrome (full nested span tree) from the submit control. */
function mirrorSubmitButtonChrome(
  rubrical: HTMLButtonElement,
  submit: HTMLButtonElement,
  label: string,
): void {
  rubrical.type = "button";
  rubrical.className = submit.className;
  rubrical.style.marginInlineEnd = "0.5rem";
  rubrical.dataset.rubricalAction = "true";
  rubrical.setAttribute("aria-label", label);

  if (submit.hasAttribute("dir")) {
    rubrical.dir = submit.dir;
  }

  rubrical.replaceChildren();
  if (submit.children.length > 0) {
    rubrical.append(cloneSubmitContent(submit, label));
  } else {
    rubrical.textContent = label;
  }

  lastMirroredSignature = structureSignature(submit);
  lastLabel = label;
}

export function updateButtonLabel(label: string): void {
  if (label === lastLabel) {
    return;
  }

  const button = document.getElementById(BUTTON_ID) as HTMLButtonElement | null;
  if (!button) {
    return;
  }

  const labelEl = findLabelElement(button);
  if (labelEl) {
    labelEl.textContent = label;
  } else {
    button.textContent = label;
  }
  button.setAttribute("aria-label", label);
  lastLabel = label;
}

export function setRubricalButtonEnabled(enabled: boolean): void {
  const button = document.getElementById(BUTTON_ID) as HTMLButtonElement | null;
  if (!button) {
    return;
  }

  button.disabled = !enabled;
  if (enabled) {
    button.removeAttribute("aria-disabled");
    button.style.opacity = "";
    button.style.cursor = "";
  } else {
    button.setAttribute("aria-disabled", "true");
    button.style.opacity = "0.55";
    button.style.cursor = "not-allowed";
  }
}

export function ensureInlineButton(
  label: string,
  onClick: (label: string) => void,
): boolean {
  const submit = findSubmitAnchor();
  if (!submit?.parentElement) {
    return false;
  }

  let button = document.getElementById(BUTTON_ID) as HTMLButtonElement | null;

  if (!button) {
    button = document.createElement("button");
    button.id = BUTTON_ID;
    button.addEventListener("click", () => {
      onClick(button!.getAttribute("aria-label") ?? label);
    });
    mirrorSubmitButtonChrome(button, submit, label);
    submit.insertAdjacentElement("beforebegin", button);
    return true;
  }

  if (button.nextElementSibling !== submit) {
    submit.insertAdjacentElement("beforebegin", button);
  }

  const structureChanged = structureSignature(submit) !== lastMirroredSignature;
  if (structureChanged) {
    mirrorSubmitButtonChrome(button, submit, label);
    return true;
  }

  updateButtonLabel(label);
  return true;
}

export function debounce<T extends (...args: never[]) => void>(
  fn: T,
  waitMs: number,
): T {
  let timer: ReturnType<typeof setTimeout> | undefined;
  return ((...args: Parameters<T>) => {
    clearTimeout(timer);
    timer = setTimeout(() => fn(...args), waitMs);
  }) as T;
}
