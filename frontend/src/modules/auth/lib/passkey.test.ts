import { describe, expect, it, vi } from "vitest";
import { base64UrlToArrayBuffer, isPasskeySupported, signInWithPasskey } from "./passkey";

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

	it("is false when isUserVerifyingPlatformAuthenticatorAvailable is missing", async () => {
		vi.stubGlobal("window", {
			PublicKeyCredential: {},
		});
		vi.stubGlobal("navigator", {
			credentials: {},
		});
		expect(await isPasskeySupported()).toBe(false);
		vi.unstubAllGlobals();
	});
});

describe("signInWithPasskey", () => {
	it("treats NotAllowedError DOMException as user cancellation", async () => {
		vi.stubGlobal("navigator", {
			credentials: {
				get: vi.fn().mockRejectedValue(new DOMException("The operation either timed out or was not allowed", "NotAllowedError")),
			},
		});
		const result = await signInWithPasskey({
			fetchOptions: async () => ({ challenge: "AQID" }),
			verify: vi.fn(),
		});
		expect(result).toEqual({ ok: false, error: "cancelled" });
		vi.unstubAllGlobals();
	});
});
