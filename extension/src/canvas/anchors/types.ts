export type CanvasAnchor = {
  a2: readonly string[];
  classic: readonly string[];
  extra?: readonly string[];
  /** Override DOM selector read for the a2 tier. */
  readA2?: () => string | null | undefined | "";
  /** Override DOM selector read for the classic tier. */
  readClassic?: () => string | null | undefined | "";
  /** Canvas `window.ENV.ASSIGNMENT` fallback — not a DOM selector. */
  env?: () => string | null | undefined | "";
};
