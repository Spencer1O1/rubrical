import { Window } from "happy-dom";
import { beforeEach, describe, expect, it } from "vitest";
import { htmlScrollLockSnapshot } from "./rubric-scroll-debug";

const dom = new Window();

describe("htmlScrollLockSnapshot", () => {
  beforeEach(() => {
    (globalThis as typeof globalThis & { document: Document }).document =
      dom.document as unknown as Document;
    document.documentElement.setAttribute("style", "");
  });

  it("reads InstUI scroll-lock inline properties from html", () => {
    document.documentElement.setAttribute(
      "style",
      "width: calc(100% - 19px); position: fixed; top: -957.5px; overflow: hidden;",
    );

    expect(htmlScrollLockSnapshot()).toEqual({
      position: "fixed",
      top: "-957.5px",
      width: "calc(100% - 19px)",
      overflow: "hidden",
    });
  });
});
