export type CanvasAssignmentEnv = {
  name?: string;
  description?: string;
  due_at?: string | null;
  points_possible?: number | null;
  submission_types?: string[] | null;
};

export type CanvasPageEnv = {
  ASSIGNMENT?: CanvasAssignmentEnv;
  can_attach_entries?: boolean | string;
};

function readCanvasPageEnv(): CanvasPageEnv | null {
  return (window as Window & { ENV?: CanvasPageEnv }).ENV ?? null;
}

export function readCanvasAssignmentEnv(): CanvasAssignmentEnv | null {
  return readCanvasPageEnv()?.ASSIGNMENT ?? null;
}

/** Course setting: Discussions → Attach files to discussions (fixtures/3-discussion*.html). */
export function readCanAttachDiscussionEntries(): boolean | undefined {
  const value = readCanvasPageEnv()?.can_attach_entries;
  if (typeof value === "boolean") {
    return value;
  }
  if (value === "true") {
    return true;
  }
  if (value === "false") {
    return false;
  }
  return undefined;
}

/** Raw ENV field readers — formatting belongs on the anchor `env` reader or `readElement`. */
export const envReaders = {
  description: (): string => readCanvasAssignmentEnv()?.description?.trim() ?? "",
  name: (): string => readCanvasAssignmentEnv()?.name?.trim() ?? "",
  dueAt: (): string => readCanvasAssignmentEnv()?.due_at?.trim() ?? "",
  pointsPossible: (): number | null | undefined => readCanvasAssignmentEnv()?.points_possible,
  submissionTypes: (): string[] => readCanvasAssignmentEnv()?.submission_types ?? [],
};
