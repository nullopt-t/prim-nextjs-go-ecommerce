import { describe, expect, it } from "vitest";
import { NOTIFY_ME_COOLDOWN_MS, isInCooldown, EMAIL_RE } from "./notifyMeStore";

describe("isInCooldown", () => {
	it("is false when there is no existing alert", () => {
		expect(isInCooldown(undefined, 1_000_000)).toBe(false);
	});

	it("is true inside the cooldown window", () => {
		expect(isInCooldown({ email: "a@example.com", at: 1_000_000 }, 1_000_000 + 1_000)).toBe(true);
	});

	it("opens up once the cooldown has passed", () => {
		expect(
			isInCooldown({ email: "a@example.com", at: 1_000_000 }, 1_000_000 + NOTIFY_ME_COOLDOWN_MS),
		).toBe(false);
	});
});

describe("EMAIL_RE gate (via subscribe validation shape)", () => {
	it("rejects obvious non-emails", () => {
		expect(EMAIL_RE.test("not-an-email")).toBe(false);
		expect(EMAIL_RE.test("@missing-user.com")).toBe(false);
		expect(EMAIL_RE.test("user@")).toBe(false);
		expect(EMAIL_RE.test("user@domain")).toBe(false);
		expect(EMAIL_RE.test("a@b.co")).toBe(true);
		expect(EMAIL_RE.test("test.user+tag@example.co.uk")).toBe(true);
	});
});
