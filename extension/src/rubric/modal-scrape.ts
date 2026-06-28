/** Shared DOM wait + modal helpers for rubric scraping (long descriptions, discussion tray). */

export function untilDom(ready: () => boolean, timeoutMs = 5000): Promise<void> {
  if (ready()) {
    return Promise.resolve();
  }

  return new Promise((resolve) => {
    const stop = () => {
      observer.disconnect();
      clearTimeout(timer);
      resolve();
    };
    const observer = new MutationObserver(() => {
      if (ready()) {
        stop();
      }
    });
    observer.observe(document.documentElement, { childList: true, subtree: true });
    const timer = setTimeout(stop, timeoutMs);
  });
}

export function modalSelector(ariaLabel: string): string {
  return `[role="dialog"][aria-label="${ariaLabel}"]`;
}

export function closeButtonIn(root: ParentNode): HTMLButtonElement | null {
  const close = root.querySelector(
    'button[data-cid*="CloseButton"], [data-cid*="CloseButton"] button, [class*="closeButton"] > button',
  );
  return close instanceof HTMLButtonElement ? close : null;
}
