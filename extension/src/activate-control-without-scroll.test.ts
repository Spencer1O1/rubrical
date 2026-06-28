import { Window } from "happy-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { beginLongDescriptionScrape } from "./activate-control-without-scroll";

const dom = new Window();

describe("beginLongDescriptionScrape", () => {
  beforeEach(() => {
    (globalThis as typeof globalThis & { document: Document; window: Window }).document =
      dom.document as unknown as Document;
    (globalThis as typeof globalThis & { window: Window }).window =
      dom.window as unknown as Window & typeof globalThis.window;
    (globalThis as typeof globalThis & { HTMLElement: typeof HTMLElement }).HTMLElement =
      dom.window.HTMLElement as unknown as typeof HTMLElement;
    Object.defineProperty(window, "scrollX", { configurable: true, value: 0 });
    Object.defineProperty(window, "scrollY", { configurable: true, value: 420 });
    document.body.setAttribute("style", "");
    document.documentElement.classList.remove("rubrical-scrape-lock");
  });

  it("locks body scroll and restores on end", () => {
    const scrollTo = vi.spyOn(window, "scrollTo");
    const end = beginLongDescriptionScrape({ lockScroll: true });

    expect(document.documentElement.classList.contains("rubrical-scrape-lock")).toBe(true);
    expect(document.body.style.position).toBe("fixed");
    expect(document.body.style.top).toBe("-420px");
    expect(document.body.style.overflowY).toBe("scroll");

    end();

    expect(document.documentElement.classList.contains("rubrical-scrape-lock")).toBe(false);
    expect(scrollTo).toHaveBeenCalledWith(0, 420);
  });

  it("skips body scroll lock when lockScroll is false", () => {
    const scrollTo = vi.spyOn(window, "scrollTo");
    const end = beginLongDescriptionScrape({ lockScroll: false });

    expect(document.documentElement.classList.contains("rubrical-scrape-lock")).toBe(true);
    expect(document.body.style.position).toBe("");
    expect(document.body.style.overflowY).toBe("");

    end();

    expect(document.documentElement.classList.contains("rubrical-scrape-lock")).toBe(false);
    expect(scrollTo).not.toHaveBeenCalled();
  });

  it("forces preventScroll on focus while active", () => {
    const button = document.createElement("button");
    document.body.append(button);
    const focus = vi.spyOn(HTMLElement.prototype, "focus");
    const end = beginLongDescriptionScrape();

    button.focus();
    expect(focus).toHaveBeenCalledWith({ preventScroll: true });

    end();
    focus.mockRestore();
  });
});
