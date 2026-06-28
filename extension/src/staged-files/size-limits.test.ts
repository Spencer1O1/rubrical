import { describe, expect, it } from "vitest";
import { isStagedFileTooLarge, MAX_STAGED_FILE_BYTES, stagedFileSizeError } from "./size-limits";

describe("size-limits", () => {
  it("flags files above the staging cap", () => {
    expect(isStagedFileTooLarge(MAX_STAGED_FILE_BYTES)).toBe(false);
    expect(isStagedFileTooLarge(MAX_STAGED_FILE_BYTES + 1)).toBe(true);
  });

  it("formats a helpful error message", () => {
    expect(stagedFileSizeError("big.pdf", MAX_STAGED_FILE_BYTES + 1)).toMatch(/500 MB/);
  });
});
