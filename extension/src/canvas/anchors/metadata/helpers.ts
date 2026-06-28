const SUBMISSION_TYPE_LABELS: Record<string, string> = {
  online_text_entry: "Online Text Entry",
  online_upload: "Online Upload",
  online_url: "Website URL",
  online_quiz: "Online Quiz",
  discussion_topic: "Discussion Topic",
  external_tool: "External Tool",
  media_recording: "Media Recording",
  student_annotation: "Student Annotation",
};

export function normalizeDueLabel(text: string): string {
  const trimmed = text.trim();
  if (!trimmed) {
    return "";
  }
  return trimmed.toLowerCase().startsWith("due") ? trimmed : `Due ${trimmed}`;
}

export function normalizePointsLabel(text: string): string {
  const trimmed = text.trim();
  if (!trimmed || !/\d/.test(trimmed)) {
    return "";
  }
  if (/pts?|points/i.test(trimmed)) {
    return trimmed;
  }
  return `${trimmed} pts`;
}

export function humanizeSubmissionType(raw: string): string {
  const key = raw.trim().toLowerCase();
  if (!key) {
    return "";
  }
  if (SUBMISSION_TYPE_LABELS[key]) {
    return SUBMISSION_TYPE_LABELS[key];
  }
  return key.replace(/_/g, " ").replace(/\b\w/g, (char) => char.toUpperCase());
}

/** Classic Canvas pages label metadata with plain text, e.g. "Submitting: a file upload". */
export function findLabeledText(label: string): string {
  for (const node of Array.from(document.querySelectorAll("span, div, dt, strong, p"))) {
    const text = node.textContent?.trim() ?? "";
    if (text.startsWith(label)) {
      return text;
    }
  }
  return "";
}
