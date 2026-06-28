import { envReaders } from "./assignment-env";
import { isStrictExtraction } from "../strict";

function decodeHtmlEntities(value: string): string {
  const trimmed = value.trim();
  if (!trimmed.includes("&lt;") && !trimmed.includes("&gt;") && !trimmed.includes("&amp;")) {
    return trimmed;
  }

  const textarea = document.createElement("textarea");
  textarea.innerHTML = trimmed;
  return textarea.value.trim();
}

export function normalizeInstructionHTML(html: string): string {
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

export function readInstructionElement(element: Element): string {
  const root =
    element.matches(".user_content, .user_content.enhanced")
      ? element
      : (element.querySelector(".user_content, .user_content.enhanced") ?? element);
  return normalizeInstructionHTML(root.innerHTML);
}

export function readInstructionEnv(): string {
  return normalizeInstructionHTML(envReaders.description());
}
