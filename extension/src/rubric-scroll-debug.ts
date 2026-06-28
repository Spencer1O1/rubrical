export type LongDescriptionScrollPhase =
  | "before-click"
  | "after-click"
  | "modal-open"
  | "before-close"
  | "after-close"
  | "after-modal-removed";

const DEBUG_STORAGE_KEY = "rubrical:debug-long-description-scroll";
const LOG_PREFIX = "[rubrical scroll]";

export function isLongDescriptionScrollDebugEnabled(): boolean {
  try {
    return localStorage.getItem(DEBUG_STORAGE_KEY) === "1";
  } catch {
    return false;
  }
}

/** Dev console: localStorage.setItem("rubrical:debug-long-description-scroll", "1") then reload. */
export function htmlScrollLockSnapshot(html: HTMLElement = document.documentElement): {
  position: string;
  top: string;
  width: string;
  overflow: string;
} {
  const { position, top, width, overflow } = html.style;
  return { position, top, width, overflow };
}

function formatScrollToArgs(args: unknown[]): string {
  if (args.length === 1 && typeof args[0] === "object" && args[0] !== null) {
    const options = args[0] as ScrollToOptions;
    return JSON.stringify({
      left: options.left,
      top: options.top,
      behavior: options.behavior,
    });
  }
  return args.map(String).join(", ");
}

function shortStack(): string {
  const stack = new Error().stack;
  if (!stack) {
    return "";
  }
  return stack
    .split("\n")
    .slice(2, 5)
    .map((line) => line.trim())
    .join(" | ");
}

export function createLongDescriptionScrollDebugSession(): {
  logPhase: (iteration: number, phase: LongDescriptionScrollPhase) => void;
  stop: () => void;
} {
  if (!isLongDescriptionScrollDebugEnabled()) {
    return { logPhase: () => {}, stop: () => {} };
  }

  let currentIteration = 0;
  let currentPhase: LongDescriptionScrollPhase = "before-click";

  const win = window;
  const originalScrollTo = win.scrollTo.bind(win);
  const originalScroll = win.scroll.bind(win);
  const elementCtor = globalThis.Element;
  const hadScrollIntoView =
    elementCtor !== undefined &&
    typeof elementCtor.prototype.scrollIntoView === "function";
  const originalScrollIntoView = hadScrollIntoView
    ? elementCtor.prototype.scrollIntoView
    : null;

  const logProgrammaticScroll = (kind: string, detail: string): void => {
    console.info(
      `${LOG_PREFIX} ${kind}(${detail}) iter=${currentIteration} phase=${currentPhase} scrollY=${win.scrollY} stack=${shortStack()}`,
    );
  };

  win.scrollTo = ((...args: unknown[]) => {
    logProgrammaticScroll("scrollTo", formatScrollToArgs(args));
    originalScrollTo(...(args as Parameters<Window["scrollTo"]>));
  }) as Window["scrollTo"];

  win.scroll = ((...args: unknown[]) => {
    logProgrammaticScroll("scroll", formatScrollToArgs(args));
    originalScroll(...(args as Parameters<Window["scroll"]>));
  }) as Window["scroll"];

  if (hadScrollIntoView && originalScrollIntoView) {
    elementCtor.prototype.scrollIntoView = function scrollIntoViewLogged(
      this: Element,
      arg?: boolean | ScrollIntoViewOptions,
    ): void {
      logProgrammaticScroll(
        "scrollIntoView",
        this === document.documentElement
          ? "<html>"
          : this === document.body
            ? "<body>"
            : this.tagName.toLowerCase(),
      );
      originalScrollIntoView.call(this, arg);
    };
  }

  const scrollTopDescriptor =
    elementCtor !== undefined
      ? Object.getOwnPropertyDescriptor(elementCtor.prototype, "scrollTop")
      : undefined;
  if (elementCtor !== undefined && scrollTopDescriptor?.set) {
    Object.defineProperty(elementCtor.prototype, "scrollTop", {
      ...scrollTopDescriptor,
      set(this: Element, value: number) {
        if (this === document.documentElement || this === document.body) {
          logProgrammaticScroll(
            "scrollTop=",
            `${value} target=${this === document.documentElement ? "<html>" : "<body>"}`,
          );
        }
        scrollTopDescriptor.set!.call(this, value);
      },
    });
  }

  console.info(
    `${LOG_PREFIX} debug enabled — disable with localStorage.removeItem("${DEBUG_STORAGE_KEY}")`,
  );

  return {
    logPhase(iteration: number, phase: LongDescriptionScrollPhase): void {
      currentIteration = iteration;
      currentPhase = phase;
      const lock = htmlScrollLockSnapshot();
      console.info(
        `${LOG_PREFIX} iter=${iteration} phase=${phase} scrollX=${win.scrollX} scrollY=${win.scrollY} html=${JSON.stringify(lock)}`,
      );
    },
    stop(): void {
      win.scrollTo = originalScrollTo;
      win.scroll = originalScroll;
      if (originalScrollIntoView) {
        elementCtor.prototype.scrollIntoView = originalScrollIntoView;
      }
      if (elementCtor !== undefined && scrollTopDescriptor) {
        Object.defineProperty(elementCtor.prototype, "scrollTop", scrollTopDescriptor);
      }
      console.info(`${LOG_PREFIX} debug stopped`);
    },
  };
}
