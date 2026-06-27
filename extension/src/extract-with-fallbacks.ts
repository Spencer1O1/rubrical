import { isStrictExtraction } from "./strict";

function hasValue<T>(value: T | null | undefined | ""): value is T {
  if (value === null || value === undefined) {
    return false;
  }
  if (typeof value === "string") {
    return value.trim() !== "";
  }
  return true;
}

/** A2 first, classic second, optional tertiary third. Strict mode stops after A2. */
export function extractWithFallbacks<T>(
  a2: () => T | null | undefined | "",
  classic: () => T | null | undefined | "",
  extra?: () => T | null | undefined | "",
): T | "" {
  const primary = a2();
  if (hasValue(primary)) {
    return primary;
  }

  if (isStrictExtraction()) {
    return "";
  }

  const secondary = classic();
  if (hasValue(secondary)) {
    return secondary;
  }

  if (extra) {
    const tertiary = extra();
    if (hasValue(tertiary)) {
      return tertiary;
    }
  }

  return "";
}

/** Same tier order for nullable results (e.g. rubric table). */
export function extractNullableWithFallbacks<T>(
  a2: () => T | null,
  classic: () => T | null,
  extra?: () => T | null,
): T | null {
  const primary = a2();
  if (primary) {
    return primary;
  }

  if (isStrictExtraction()) {
    return null;
  }

  const secondary = classic();
  if (secondary) {
    return secondary;
  }

  return extra?.() ?? null;
}
