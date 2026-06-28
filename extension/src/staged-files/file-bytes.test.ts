import { describe, expect, it } from "vitest";
import { arrayBufferToBase64, base64ToArrayBuffer, coerceToArrayBuffer } from "./file-bytes";

describe("file bytes", () => {
  it("round-trips base64", () => {
    const original = new Uint8Array([37, 80, 68, 70, 45, 49, 46, 52]).buffer;
    const restored = base64ToArrayBuffer(arrayBufferToBase64(original));
    expect(Array.from(new Uint8Array(restored))).toEqual([37, 80, 68, 70, 45, 49, 46, 52]);
  });

  it("accepts ArrayBuffer in coerceToArrayBuffer", () => {
    const buffer = new Uint8Array([37, 80, 68, 70]).buffer;
    expect(Array.from(new Uint8Array(coerceToArrayBuffer(buffer)))).toEqual([37, 80, 68, 70]);
  });

  it("accepts Uint8Array in coerceToArrayBuffer", () => {
    const view = new Uint8Array([37, 80, 68, 70]);
    expect(Array.from(new Uint8Array(coerceToArrayBuffer(view)))).toEqual([37, 80, 68, 70]);
  });

  it("rejects plain objects that would become [object Object]", () => {
    expect(() => coerceToArrayBuffer({})).toThrow(/ArrayBuffer/);
  });
});
