import { describe, expect, it } from "vitest";
import { NOTIFY_ME_COOLDOWN_MS, isInCooldown } from "./notifyMeStore";

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
		expect(/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test("not-an-email")).toBe(false);
		expect(/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test("a@b.co")).toBe(true);
	});
});
