"use client";

/**
 * Passkey (WebAuthn) sign-in. Forward-compatible: it talks to
 * `/api/v1/auth/passkey/authentication/options|verify` when the backend
 * enables them; until then the entry point degrades gracefully (button
 * hidden when the platform authenticator is unavailable).
 */

export interface PasskeyOptions {
	challenge: string;
	rpId?: string;
	userVerification?: "preferred" | "required" | "discouraged";
}

export interface PasskeyVerifyResult {
	ok: boolean;
	error?: string;
}

/** Narrow assertion shape (TS DOM lib predates AuthenticatorAssertionCredential). */
export interface PasskeyAssertion {
	id: string;
	type: string;
	rawId: ArrayBuffer;
	response: AuthenticatorAssertionResponse;
}

/** Decodes a base64url string (WebAuthn challenge format) into bytes. */
export function base64UrlToArrayBuffer(value: string): ArrayBuffer {
	const padding = "=".repeat((4 - (value.length % 4)) % 4);
	const base64 = (value + padding).replace(/-/g, "+").replace(/_/g, "/");
	const binary = atob(base64);
	const bytes = new Uint8Array(binary.length);
	for (let i = 0; i < binary.length; i += 1) bytes[i] = binary.charCodeAt(i);
	return bytes.buffer;
}

export async function isPasskeySupported(): Promise<boolean> {
	if (typeof window === "undefined") return false;
	if (!("credentials" in navigator) || typeof window.PublicKeyCredential === "undefined") {
		return false;
	}
	try {
		return await PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable();
	} catch {
		return false;
	}
}

/**
 * Runs the full authentication ceremony. `fetchOptions` and `verify` are
 * injected by the component so this stays free of the HTTP client and can
 * be tested in isolation.
 */
export async function signInWithPasskey(deps: {
	fetchOptions: () => Promise<PasskeyOptions>;
	verify: (assertion: PasskeyAssertion, requestId?: string) => Promise<PasskeyVerifyResult>;
}): Promise<PasskeyVerifyResult> {
	const options = await deps.fetchOptions();
	const credential = (await navigator.credentials.get({
		publicKey: {
			challenge: base64UrlToArrayBuffer(options.challenge),
			rpId: options.rpId,
			userVerification: options.userVerification ?? "preferred",
		},
	})) as PasskeyAssertion | null;
	if (!credential) return { ok: false, error: "cancelled" };
	const requestId = (options as { requestId?: string }).requestId;
	return deps.verify(credential, requestId);
}
