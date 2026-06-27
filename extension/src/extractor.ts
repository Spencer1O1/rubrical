import { extractWithFallbacks } from "./extract-with-fallbacks";
import { isStrictExtraction } from "./strict";

export function findLabeledText(label: string): string {
  const nodes = Array.from(document.querySelectorAll("span, div, dt, strong, p"));
  for (const node of nodes) {
    const text = node.textContent?.trim() ?? "";
    if (text.startsWith(label)) {
      return text;
    }
  }
  return "";
}

type CanvasAssignmentEnv = {
  name?: string;
  description?: string;
};

function readEnvAssignment(): CanvasAssignmentEnv | null {
  const assignment = (window as Window & { ENV?: { ASSIGNMENT?: CanvasAssignmentEnv } }).ENV
    ?.ASSIGNMENT;
  return assignment ?? null;
}

function normalizeLabel(value: string | null | undefined): string {
  return (value ?? "").replace(/\s+/g, " ").trim();
}

function decodeHtmlEntities(value: string): string {
  const trimmed = value.trim();
  if (!trimmed.includes("&lt;") && !trimmed.includes("&gt;") && !trimmed.includes("&amp;")) {
    return trimmed;
  }

  const textarea = document.createElement("textarea");
  textarea.innerHTML = trimmed;
  return textarea.value.trim();
}

function normalizeInstructionHTML(html: string): string {
  const trimmed = html.trim();
  if (!trimmed) {
    return "";
  }

  const decoded = isStrictExtraction() ? trimmed : decodeHtmlEntities(trimmed);
  const template = document.createElement("template");
  template.innerHTML = decoded;
  const userContent = template.content.querySelector(".user_content");
  if (userContent?.innerHTML.trim()) {
    return userContent.innerHTML.trim();
  }

  return decoded;
}

function readInstructionsHTMLFromElement(el: Element | null): string {
  if (!el) {
    return "";
  }

  const userContent = el.querySelector(".user_content, .user_content.enhanced");
  if (userContent?.innerHTML.trim()) {
    return normalizeInstructionHTML(userContent.innerHTML);
  }

  return normalizeInstructionHTML(el.innerHTML);
}

function findA2DetailsPanel(): Element | null {
  const buttons = Array.from(
    document.querySelectorAll<HTMLButtonElement>(
      '[data-testid="assignments-2-student-view"] button[class*="toggleDetails"], #assignment_show button[id^="Expandable"], #assignment_show button[class*="toggleDetails"]',
    ),
  );

  for (const button of buttons) {
    if (normalizeLabel(button.textContent) !== "Details") {
      continue;
    }

    const controlsId = button.getAttribute("aria-controls");
    if (controlsId) {
      const panel = document.getElementById(controlsId);
      if (panel?.textContent?.trim()) {
        return panel;
      }
    }

    const panel = button.parentElement?.querySelector(
      '[class*="toggleDetails__details"], [class*="toggleDetails__content"]',
    );
    if (panel?.textContent?.trim()) {
      return panel;
    }
  }

  return null;
}

function extractA2Instructions(): string {
  const fromDetails = readInstructionsHTMLFromElement(findA2DetailsPanel());
  if (fromDetails) {
    return fromDetails;
  }

  return readInstructionsHTMLFromElement(
    document.querySelector('[data-testid="assignment-description"], [data-testid="assignment-instructions"]'),
  );
}

function extractClassicInstructions(): string {
  return readInstructionsHTMLFromElement(
    document.querySelector("#assignment_show .description .user_content, #assignment_show .user_content"),
  );
}

function extractEnvInstructions(): string {
  const description = readEnvAssignment()?.description?.trim() ?? "";
  return normalizeInstructionHTML(description);
}

/** Assignment instructions HTML: A2 → classic → ENV (tertiary). */
export function extractInstructions(): string {
  return (
    extractWithFallbacks(extractA2Instructions, extractClassicInstructions, extractEnvInstructions) ||
    ""
  );
}

function extractA2Title(): string {
  return (
    document.querySelector(
      '[data-testid="assignments-2-student-view"] h1, #assignment_show h1',
    )?.textContent?.trim() ?? ""
  );
}

function extractClassicTitle(): string {
  return document.querySelector(".assignment-title")?.textContent?.trim() ?? "";
}

function extractExtraTitle(): string {
  return readEnvAssignment()?.name?.trim() ?? document.title.trim();
}

/** Assignment title: A2 → classic → ENV / document.title (tertiary). */
export function extractTitle(): string {
  return extractWithFallbacks(extractA2Title, extractClassicTitle, extractExtraTitle) || document.title;
}

export function extractCourseName(): string {
  return (
    document.querySelector("#breadcrumb li:last-child, .ellipsible")?.textContent?.trim() ?? ""
  );
}
