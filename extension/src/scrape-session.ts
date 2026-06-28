const sessionListeners = new Set<(active: boolean) => void>();

let sessionDepth = 0;

function notifySessionActive(active: boolean): void {
  for (const listener of sessionListeners) {
    listener(active);
  }
}

export function onLongDescriptionScrapeSession(
  listener: (active: boolean) => void,
): () => void {
  sessionListeners.add(listener);
  return () => {
    sessionListeners.delete(listener);
  };
}

export async function withLongDescriptionScrapeSession<T>(
  fn: () => Promise<T>,
): Promise<T> {
  const wasIdle = sessionDepth === 0;
  sessionDepth++;
  if (wasIdle) {
    notifySessionActive(true);
  }

  try {
    return await fn();
  } finally {
    sessionDepth--;
    if (sessionDepth === 0) {
      notifySessionActive(false);
    }
  }
}
