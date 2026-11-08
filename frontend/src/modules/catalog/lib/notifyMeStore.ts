"use client";

const STORAGE_KEY = "prim:stock-alerts";
export const NOTIFY_ME_COOKIE_PREFIX = "prim_notify_";

/** Minimum ms between subscribe attempts for the same product. */
export const NOTIFY_ME_COOLDOWN_MS = 60_000;

export interface StockAlert {
	email: string;
	/** Epoch ms of the subscription. */
	at: number;
}

export type StockAlertMap = Record<string, StockAlert>;

export const EMAIL_RE = /^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+$/;

/** Pure: true when a new attempt must wait out the cooldown. */
export function isInCooldown(existing: StockAlert | undefined, now: number): boolean {
	if (!existing) return false;
	return now - existing.at < NOTIFY_ME_COOLDOWN_MS;
}

function getCookieAlert(slug: string): StockAlert | null {
	if (typeof window === "undefined" || !slug) return null;
	try {
		const cookieName = `${NOTIFY_ME_COOKIE_PREFIX}${encodeURIComponent(slug)}`;
		const match = document.cookie.match(new RegExp(`(?:^|; )${cookieName}=([^;]*)`));
		if (!match?.[1]) return null;
		const decoded = decodeURIComponent(match[1]);
		const [email, atStr] = decoded.split("|");
		const at = Number(atStr);
		if (email && Number.isFinite(at)) {
			return { email, at };
		}
	} catch {
		// Ignore cookie read failure
	}
	return null;
}

function setCookieAlert(slug: string, alert: StockAlert) {
	if (typeof window === "undefined" || !slug) return;
	try {
		const cookieName = `${NOTIFY_ME_COOKIE_PREFIX}${encodeURIComponent(slug)}`;
		const val = encodeURIComponent(`${alert.email}|${alert.at}`);
		const maxAge = 60 * 60 * 24 * 365;
		const secure = window.location.protocol === "https:" ? "; Secure" : "";
		document.cookie = `${cookieName}=${val}; Path=/; Max-Age=${maxAge}; SameSite=Lax${secure}`;
	} catch {
		// Ignore cookie write failure
	}
}

function readMap(): StockAlertMap {
	if (typeof window === "undefined") return {};
	try {
		const raw = window.localStorage.getItem(STORAGE_KEY);
		const parsed = raw ? JSON.parse(raw) : {};
		if (!parsed || typeof parsed !== "object") return {};
		const out: StockAlertMap = {};
		for (const [slug, alert] of Object.entries(parsed as Record<string, unknown>)) {
			if (
				typeof slug === "string" &&
				slug &&
				alert &&
				typeof alert === "object" &&
				typeof (alert as StockAlert).email === "string" &&
				typeof (alert as StockAlert).at === "number"
			) {
				out[slug] = alert as StockAlert;
			}
		}
		return out;
	} catch {
		return {};
	}
}

export function getStockAlert(slug?: string | null): StockAlert | null {
	if (!slug) return null;
	const fromStorage = readMap()[slug];
	if (fromStorage) return fromStorage;
	return getCookieAlert(slug);
}

export function subscribeStockAlert(
	slug: string,
	email: string,
	now: number = Date.now(),
): { ok: boolean; reason?: "invalid-email" | "cooldown" | "already-subscribed" } {
	const trimmed = email.trim();
	if (!EMAIL_RE.test(trimmed)) return { ok: false, reason: "invalid-email" };
	const existing = getStockAlert(slug);
	if (existing && existing.email.toLowerCase() === trimmed.toLowerCase()) {
		return { ok: false, reason: "already-subscribed" };
	}
	if (isInCooldown(existing ?? undefined, now)) return { ok: false, reason: "cooldown" };
	const alert: StockAlert = { email: trimmed, at: now };
	setCookieAlert(slug, alert);
	try {
		const map = readMap();
		map[slug] = alert;
		window.localStorage.setItem(STORAGE_KEY, JSON.stringify(map));
	} catch {
		// Storage unavailable — cookie still persists.
	}
	return { ok: true };
}
