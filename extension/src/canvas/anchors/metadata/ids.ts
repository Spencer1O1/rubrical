/**
 * Metadata field test ids and classic selectors (due date, points, submission type, scope).
 */
export const metadataIds = {
  dueDate: "due-date",
  gradeDisplay: "grade-display",
} as const;

export const metadataClassic = {
  assignmentShow: "#assignment_show",
  submissionForm: ".submission_form",
  assignmentDatesDueDate: ".assignment_dates .due_date",
  assignmentShowDateDue: "#assignment_show .date_due",
  pointsPossible: ".points_possible",
  assignmentPoints: ".assignment_points",
  pointsValue: ".points-value",
  screenReaderContent: '[class*="screenReaderContent"]',
} as const;
