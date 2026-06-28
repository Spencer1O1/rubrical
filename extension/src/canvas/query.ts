import { isStrictExtraction } from "../strict";
import type { CanvasAnchor } from "./anchors/types";

function tierHasValue<T>(value: T | null | undefined | ""): value is T {
  if (value === null || value === undefined) {
    return false;
  }
  if (typeof value === "string") {
    return value.trim() !== "";
  }
  return true;
}

/** Tier runner — strict mode stops after the first tier. */
export function runTiers<T>(tiers: ReadonlyArray<() => T | null | undefined | "">): T | null {
  if (tiers.length === 0) {
    return null;
  }

  const primary = tiers[0]!();
  if (tierHasValue(primary)) {
    return primary;
  }

  if (isStrictExtraction()) {
    return null;
  }

  for (let index = 1; index < tiers.length; index++) {
    const value = tiers[index]!();
    if (tierHasValue(value)) {
      return value;
    }
  }

  return null;
}

export function testId(id: string): string {
  return `[data-testid="${id}"]`;
}

export function testIdStartsWith(prefix: string): string {
  return `[data-testid^="${prefix}"]`;
}

export function testIdContains(substr: string): string {
  return `[data-testid*="${substr}"]`;
}

export function firstMatch<T extends Element = Element>(
  selectors: readonly string[],
  root: ParentNode = document,
): T | null {
  for (const selector of selectors) {
    const match = root.querySelector<T>(selector);
    if (match) {
      return match;
    }
  }
  return null;
}

export function firstMatchAll<T extends Element = Element>(
  selectors: readonly string[],
  root: ParentNode = document,
): T[] {
  const seen = new Set<T>();
  const results: T[] = [];
  for (const selector of selectors) {
    for (const match of Array.from(root.querySelectorAll<T>(selector))) {
      if (!seen.has(match)) {
        seen.add(match);
        results.push(match);
      }
    }
  }
  return results;
}

function readFromSelectors(
  selectors: readonly string[],
  readElement: (element: Element) => string,
  root: ParentNode,
): string {
  const element = firstMatch(selectors, root);
  return element ? readElement(element) : "";
}

function domTiers<T extends Element>(
  anchor: CanvasAnchor,
  root: ParentNode,
): Array<() => T | null> {
  return [
    () => firstMatch<T>(anchor.a2, root),
    () => firstMatch<T>(anchor.classic, root),
    ...(anchor.extra ? [() => firstMatch<T>(anchor.extra!, root)] : []),
  ];
}

/** Find an element: a2 → classic → extra. Ignores `env` and custom `read*` fns. */
export function queryAnchor<T extends Element = Element>(
  anchor: CanvasAnchor,
  root: ParentNode = document,
): T | null {
  return runTiers(domTiers<T>(anchor, root));
}

export function queryAnchorAll<T extends Element = Element>(
  anchor: CanvasAnchor,
  root: ParentNode = document,
): T[] {
  const tiers = isStrictExtraction()
    ? [anchor.a2]
    : [anchor.a2, anchor.classic, ...(anchor.extra ? [anchor.extra] : [])];
  const seen = new Set<T>();
  const results: T[] = [];
  for (const selectors of tiers) {
    for (const match of firstMatchAll<T>(selectors, root)) {
      if (!seen.has(match)) {
        seen.add(match);
        results.push(match);
      }
    }
  }
  return results;
}

export function anyAnchorPresent(anchor: CanvasAnchor, root: ParentNode = document): boolean {
  return queryAnchorAll(anchor, root).length > 0;
}

export function combinedSelector(anchor: CanvasAnchor): string {
  return [...anchor.a2, ...anchor.classic, ...(anchor.extra ?? [])].join(", ");
}

/**
 * Extract a string: a2 → classic → extra → env.
 * Custom `readA2` / `readClassic` override selector reads; default read is textContent.
 */
export function extractAnchor(
  anchor: CanvasAnchor,
  readElement: (element: Element) => string = (element) => element.textContent?.trim() ?? "",
  root: ParentNode = document,
): string {
  const tiers: Array<() => string | null | undefined | ""> = [
    () => anchor.readA2?.() ?? readFromSelectors(anchor.a2, readElement, root),
    () => anchor.readClassic?.() ?? readFromSelectors(anchor.classic, readElement, root),
  ];

  if (anchor.extra) {
    tiers.push(() => readFromSelectors(anchor.extra!, readElement, root));
  }

  if (anchor.env) {
    tiers.push(anchor.env);
  }

  return runTiers(tiers) ?? "";
}
