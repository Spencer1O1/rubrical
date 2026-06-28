import { instructions, title } from "./canvas/anchors";
import { extractAnchor, firstMatch } from "./canvas/query";
import { readInstructionElement } from "./canvas/instruction-html";

export function extractInstructions(): string {
  return extractAnchor(instructions.description, readInstructionElement);
}

export function extractTitle(): string {
  return extractAnchor(title.heading) || document.title;
}

export function extractCourseName(): string {
  return extractAnchor(title.breadcrumb);
}
