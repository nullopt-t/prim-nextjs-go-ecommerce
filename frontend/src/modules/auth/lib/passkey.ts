"use client";

/**
 * Passkey (WebAuthn) sign-in. Forward-compatible: it talks to
 * `/api/v1/auth/passkey/authentication/options|verify` when the backend
 * enables them; until then the entry point degrades gracefully (button
 * hidden when the platform authenticator is unavailable).
 */

export interface PasskeyOptions {
	requestId?: string;
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

/** Decodes a base64url string (WebAuthn challenge format) into bytes. Safe across browser and SSR/Node runtimes. */
export function base64UrlToArrayBuffer(value: string): ArrayBuffer {
	if (typeof window === "undefined" && typeof Buffer !== "undefined") {
		const buf = Buffer.from(value, "base64url");
		return buf.buffer.slice(buf.byteOffset, buf.byteOffset + buf.byteLength);
	}
	const padding = "=".repeat((4 - (value.length % 4)) % 4);
	const base64 = (value + padding).replace(/-/g, "+").replace(/_/g, "/");
	const binary = typeof atob === "function" ? atob(base64) : Buffer.from(base64, "base64").toString("binary");
	const bytes = new Uint8Array(binary.length);
	for (let i = 0; i < binary.length; i += 1) bytes[i] = binary.charCodeAt(i);
	return bytes.buffer;
}

export async function isPasskeySupported(): Promise<boolean> {
	if (typeof window === "undefined") return false;
	if (!("credentials" in navigator) || typeof navigator.credentials?.get !== "function") return false;
	if (typeof window.PublicKeyCredential === "undefined") return false;
	const pkc = window.PublicKeyCredential as typeof PublicKeyCredential & {
		isConditionalMediationAvailable?: unknown;
	};
	if (typeof pkc.isUserVerifyingPlatformAuthenticatorAvailable === "function") {
		try {
			if (await pkc.isUserVerifyingPlatformAuthenticatorAvailable()) return true;
		} catch {
			// Fall through to roaming-authenticator fallback.
		}
	}
	if (typeof pkc.isConditionalMediationAvailable === "function") {
		try {
			if (await (pkc.isConditionalMediationAvailable as () => Promise<boolean>)()) return true;
		} catch {
			// Fall through to roaming-authenticator fallback.
		}
	}
	// Fallback: WebAuthn API present (roaming/security-key/phone passkeys).
	return true;
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
	let credential: PasskeyAssertion | null = null;
	try {
		credential = (await navigator.credentials.get({
			publicKey: {
				challenge: base64UrlToArrayBuffer(options.challenge),
				rpId: options.rpId,
				userVerification: options.userVerification ?? "preferred",
			},
		})) as PasskeyAssertion | null;
	} catch (err) {
		if (err instanceof DOMException && (err.name === "NotAllowedError" || err.name === "AbortError")) {
			return { ok: false, error: "cancelled" };
		}
		throw err;
	}
	if (!credential) return { ok: false, error: "cancelled" };
	const requestId: string | undefined = options.requestId ? String(options.requestId) : undefined;
	return deps.verify(credential, requestId);
}
