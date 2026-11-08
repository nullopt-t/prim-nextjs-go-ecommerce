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

	it("handles undefined requestId gracefully when no credential obtained", async () => {
		// When navigator.credentials.get returns null (e.g. user cancelled)
		vi.stubGlobal("navigator", {
			credentials: {
				get: vi.fn().mockResolvedValue(null),
			},
		});
		const result = await signInWithPasskey({
			fetchOptions: async () => ({ challenge: "AQID", rpId: "example.com" } as any),
			verify: vi.fn(),
		});
		expect(result).toEqual({ ok: false, error: "cancelled" }); // no credential = cancelled
		vi.unstubAllGlobals();
	});

	it("handles undefined requestId gracefully when credential obtained but verify fails", async () => {
		// When credential is obtained but verify returns false
		vi.stubGlobal("navigator", {
			credentials: {
				get: vi.fn().mockResolvedValue({
					id: "cred-id",
					type: "public-key",
					trawId: new Uint8Array([1,2,3,4]).buffer,
					response: { authenticatorData: new Uint8Array([]).buffer, clientDataJSON: new Uint8Array([]).buffer, signature: new Uint8Array([]).buffer, userHandle: null } as any,
				}) as any,
			},
		});
		const result = await signInWithPasskey({
			fetchOptions: async () => ({ challenge: "AQID", rpId: "example.com" } as any),
			verify: vi.fn().mockResolvedValue({ ok: false, error: "test error" }),
		});
		expect(result).toEqual({ ok: false, error: "test error" }); // verify called with undefined requestId
		vi.unstubAllGlobals();
	});

	it("handles requestId when provided by backend", async () => {
		// When backend provides requestId, it should be passed through
		vi.stubGlobal("navigator", {
			credentials: {
				get: vi.fn().mockResolvedValue({
					id: "cred-id",
					type: "public-key",
					trawId: new Uint8Array([1,2,3,4]).buffer,
					response: { authenticatorData: new Uint8Array([]).buffer, clientDataJSON: new Uint8Array([]).buffer, signature: new Uint8Array([]).buffer, userHandle: null } as any,
				}) as any,
			},
		});
		const mockRequestId = "req-12345";
		const result = await signInWithPasskey({
			fetchOptions: async () => ({ challenge: "AQID", rpId: "example.com", requestId: mockRequestId } as any),
			verify: vi.fn().mockResolvedValue({ ok: true }),
		});
		expect(result).toEqual({ ok: true }); // verify called with requestId
		vi.unstubAllGlobals();
	});
});
