import { describe, expect, it } from "vitest";
import { isSafeUrl } from "./safeUrl";

describe("isSafeUrl", () => {
	it("accepts http and https URLs", () => {
		expect(isSafeUrl("http://carrier.test/track/1")).toBe(true);
		expect(isSafeUrl("https://carrier.test/track/1")).toBe(true);
	});

	it("rejects executable and non-web schemes", () => {
		expect(isSafeUrl("javascript:alert(1)")).toBe(false);
		expect(isSafeUrl("JaVaScRiPt:alert(1)")).toBe(false);
		expect(isSafeUrl("data:text/html;base64,PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg==")).toBe(false);
		expect(isSafeUrl("//carrier.test/track/1")).toBe(false);
		expect(isSafeUrl("/relative/path")).toBe(false);
	});

	it("rejects empty and nullish values", () => {
		expect(isSafeUrl("")).toBe(false);
		expect(isSafeUrl("   ")).toBe(false);
		expect(isSafeUrl(null)).toBe(false);
		expect(isSafeUrl(undefined)).toBe(false);
	});
});
