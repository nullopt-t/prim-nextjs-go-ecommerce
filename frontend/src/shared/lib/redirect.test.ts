import { describe, expect, it } from "vitest";
import { safeRedirectPath } from "./redirect";

describe("safeRedirectPath", () => {
  it("allows internal paths", () => {
    expect(safeRedirectPath("/vendor")).toBe("/vendor");
    expect(safeRedirectPath("/en/user/orders?page=2")).toBe("/en/user/orders?page=2");
  });

  it("rejects protocol-relative and absolute URLs", () => {
    expect(safeRedirectPath("//evil.example")).toBe("/");
    expect(safeRedirectPath("/\\evil.example")).toBe("/");
    expect(safeRedirectPath("https://evil.example")).toBe("/");
    expect(safeRedirectPath("javascript:alert(1)")).toBe("/");
  });

  it("rejects encoded traversal and whitespace tricks", () => {
    expect(safeRedirectPath(" /vendor")).toBe("/");
    expect(safeRedirectPath("/vendor\nSet-Cookie: x=1")).toBe("/");
  });

  it("falls back for missing values", () => {
    expect(safeRedirectPath(null)).toBe("/");
    expect(safeRedirectPath("")).toBe("/");
  });
});
