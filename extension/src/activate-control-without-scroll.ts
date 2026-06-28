const STYLE_ID = "rubrical-long-description-scrape";
const LOCK_CLASS = "rubrical-scrape-lock";

const SCRAPE_CSS = `
html.${LOCK_CLASS}{position:static!important;top:auto!important;width:auto!important;overflow:unset!important}
html.${LOCK_CLASS} [aria-label="Criterion Long Description"],
html.${LOCK_CLASS} [role="dialog"][aria-label="Criterion Long Description"],
html.${LOCK_CLASS} [class*="mask"]:has([aria-label="Criterion Long Description"]){display:none!important;visibility:hidden!important;opacity:0!important;pointer-events:none!important}
html.${LOCK_CLASS} [data-position-content="discussion-post-menu"],
html.${LOCK_CLASS} [data-testid="discussion-thread-menuitem-rubric"],
html.${LOCK_CLASS} [role="menu"]:has([data-testid="discussion-thread-menuitem-rubric"]){display:none!important;visibility:hidden!important;opacity:0!important;pointer-events:none!important}
html.${LOCK_CLASS} [data-testid="assignment-rubric-modal"],
html.${LOCK_CLASS} [role="dialog"][aria-label="Assignment Rubric Details"]{visibility:hidden!important;opacity:0!important;pointer-events:none!important}
html.${LOCK_CLASS} [data-cid="Tray"]:has([data-testid="enhanced-rubric-assessment-tray"]),
html.${LOCK_CLASS} [data-testid="enhanced-rubric-assessment-tray"],
html.${LOCK_CLASS} [role="dialog"][aria-label="Rubric Assessment Tray"]{visibility:hidden!important;opacity:0!important;pointer-events:none!important}
html.${LOCK_CLASS} [data-cid="Modal"]:has([data-testid="assignment-rubric-modal"]){visibility:hidden!important;opacity:0!important;pointer-events:none!important}
`;

export function activateControlWithoutScroll(element: HTMLElement): void {
  element.focus({ preventScroll: true });
  element.dispatchEvent(
    new MouseEvent("click", { bubbles: true, cancelable: true, view: window }),
  );
}

export type BeginLongDescriptionScrapeOptions = {
  /** Freeze page scroll while scraping. Off on discussion pages (tray UI handles its own scroll). */
  lockScroll?: boolean;
};

/** Hide scrape-time overlays; optionally freeze scroll (assignment long-description modals). */
export function beginLongDescriptionScrape(
  options: BeginLongDescriptionScrapeOptions = {},
): () => void {
  const lockScroll = options.lockScroll ?? true;
  const html = document.documentElement;
  const body = document.body;
  const x = window.scrollX;
  const y = window.scrollY;
  const saved = lockScroll
    ? {
        position: body.style.position,
        top: body.style.top,
        left: body.style.left,
        right: body.style.right,
        overflowY: body.style.overflowY,
      }
    : null;
  const nativeFocus = HTMLElement.prototype.focus;
  HTMLElement.prototype.focus = function (this: HTMLElement, focusOptions?: FocusOptions) {
    nativeFocus.call(this, { ...focusOptions, preventScroll: true });
  };

  if (!document.getElementById(STYLE_ID)) {
    const style = document.createElement("style");
    style.id = STYLE_ID;
    style.textContent = SCRAPE_CSS;
    html.append(style);
  }

  html.classList.add(LOCK_CLASS);
  if (lockScroll && saved) {
    body.style.position = "fixed";
    body.style.top = `-${y}px`;
    body.style.left = "0";
    body.style.right = "0";
    body.style.overflowY = "scroll";
  }

  return () => {
    HTMLElement.prototype.focus = nativeFocus;
    html.classList.remove(LOCK_CLASS);
    if (lockScroll && saved) {
      body.style.position = saved.position;
      body.style.top = saved.top;
      body.style.left = saved.left;
      body.style.right = saved.right;
      body.style.overflowY = saved.overflowY;
      window.scrollTo(x, y);
    }
  };
}
