"use client";

const STORAGE_KEY = "prim:stock-alerts";

/** Minimum ms between subscribe attempts for the same product. */
export const NOTIFY_ME_COOLDOWN_MS = 60_000;

export interface StockAlert {
	email: string;
	/** Epoch ms of the subscription. */
	at: number;
}

export type StockAlertMap = Record<string, StockAlert>;

export const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

/** Pure: true when a new attempt must wait out the cooldown. */
export function isInCooldown(existing: StockAlert | undefined, now: number): boolean {
	if (!existing) return false;
	return now - existing.at < NOTIFY_ME_COOLDOWN_MS;
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
	return readMap()[slug] ?? null;
}

export function subscribeStockAlert(
	slug: string,
	email: string,
	now: number = Date.now(),
): { ok: boolean; reason?: "invalid-email" | "cooldown" | "already-subscribed" } {
	const trimmed = email.trim();
	if (!EMAIL_RE.test(trimmed)) return { ok: false, reason: "invalid-email" };
	const map = readMap();
	const existing = map[slug];
	if (existing && existing.email.toLowerCase() === trimmed.toLowerCase()) {
		return { ok: false, reason: "already-subscribed" };
	}
	if (isInCooldown(existing, now)) return { ok: false, reason: "cooldown" };
	try {
		map[slug] = { email: trimmed, at: now };
		window.localStorage.setItem(STORAGE_KEY, JSON.stringify(map));
	} catch {
		// Storage unavailable — still report success so the UX degrades quietly.
	}
	return { ok: true };
}
