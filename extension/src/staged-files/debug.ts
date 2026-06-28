export function rubricalDebugEnabled(): boolean {
  try {
    return localStorage.getItem("rubrical_debug") === "1";
  } catch {
    return false;
  }
}

export function rubricalDebugLog(label: string, data: Record<string, unknown>): void {
  if (!rubricalDebugEnabled()) {
    return;
  }
  console.info(`[rubrical] ${label}`, data);
}
