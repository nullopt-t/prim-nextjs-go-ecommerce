import { describe, expect, it } from "vitest";
import { normalizeSlug, validateSlug } from "./slug";

describe("seller slug policy", () => {
  it("normalizes NFKC, whitespace, and case", () => {
    expect(normalizeSlug("  ＡＢＣ  ")).toBe("abc");
  });

  it("accepts a canonical slug", () => {
    expect(validateSlug("Amrs-Gadgets")).toEqual({ slug: "amrs-gadgets" });
  });

  it("rejects reserved words", () => {
    expect(validateSlug("Admin").reason).toMatch(/reserved/i);
  });

  it("rejects short, long, and malformed slugs", () => {
    expect(validateSlug("ab").reason).toMatch(/3 and 63/);
    expect(validateSlug("a".repeat(64)).reason).toMatch(/3 and 63/);
    expect(validateSlug("my--store").reason).toMatch(/lowercase/i);
    expect(validateSlug("my_store").reason).toMatch(/lowercase/i);
  });
});
