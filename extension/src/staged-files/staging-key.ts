export type StagingKey = string;

export type StagingKeyKind = "assignment" | "discussion";

export function normalizeFileName(name: string): string {
  const base = name.split(/[/\\]/).pop() ?? name;
  return base.trim().toLowerCase().replace(/\s+/g, " ");
}

export function provisionalStorageKey(normalizedFileName: string, stagedAt: string): string {
  return `provisional:${normalizedFileName}:${stagedAt}`;
}

export function canvasStorageKey(canvasFileId: string): string {
  return `canvas:${canvasFileId}`;
}

type PageEnv = {
  COURSE_ID?: string;
  ASSIGNMENT_ID?: string;
  discussion_topic_id?: string | number;
};

function readPageEnv(): PageEnv | undefined {
  return (window as Window & { ENV?: PageEnv }).ENV;
}

function stagingKeyFromPath(pathname: string): StagingKey | null {
  const assignmentMatch = pathname.match(/\/courses\/(\d+)\/assignments\/(\d+)/);
  if (assignmentMatch) {
    return `${assignmentMatch[1]}:assignment:${assignmentMatch[2]}`;
  }

  const discussionMatch = pathname.match(/\/courses\/(\d+)\/discussion_topics\/(\d+)/);
  if (discussionMatch) {
    return `${discussionMatch[1]}:discussion:${discussionMatch[2]}`;
  }

  return null;
}

/** IndexedDB / manifest scope for the current Canvas assignment or discussion page. */
export function stagingKeyFromPage(): StagingKey | null {
  const fromPath = stagingKeyFromPath(window.location.pathname);
  if (fromPath) {
    return fromPath;
  }

  const env = readPageEnv();
  if (env?.COURSE_ID && env?.ASSIGNMENT_ID) {
    return `${env.COURSE_ID}:assignment:${env.ASSIGNMENT_ID}`;
  }

  if (env?.COURSE_ID && env.discussion_topic_id) {
    return `${env.COURSE_ID}:discussion:${env.discussion_topic_id}`;
  }

  return null;
}

export function stagingKeyFromSourceUrl(sourceUrl: string): StagingKey | null {
  try {
    return stagingKeyFromPath(new URL(sourceUrl).pathname);
  } catch {
    return null;
  }
}

export function stagingKeyKind(key: StagingKey): StagingKeyKind {
  return key.includes(":discussion:") ? "discussion" : "assignment";
}
