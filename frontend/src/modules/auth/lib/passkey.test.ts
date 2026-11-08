import { describe, expect, it, vi } from "vitest";
import { base64UrlToArrayBuffer, isPasskeySupported } from "./passkey";

describe("base64UrlToArrayBuffer", () => {
	it("decodes base64url challenges to bytes", () => {
		const decoded = new Uint8Array(base64UrlToArrayBuffer("AQID")); // 0x01 0x02 0x03
		expect([...decoded]).toEqual([1, 2, 3]);
		const padded = new Uint8Array(base64UrlToArrayBuffer("AQIDBA")); // 4 bytes → needs "=="
		expect([...padded]).toEqual([1, 2, 3, 4]);
	});
});

describe("isPasskeySupported", () => {
	it("is false outside the browser", async () => {
		expect(await isPasskeySupported()).toBe(false);
	});

	it("is false when the API is absent", async () => {
		vi.stubGlobal("navigator", {});
		vi.unstubAllGlobals();
		expect(await isPasskeySupported()).toBe(false);
	});
});
