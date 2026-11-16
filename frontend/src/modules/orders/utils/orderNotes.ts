/**
 * Order status-history notes are written by the backend as internal audit text.
 * Only free-form admin/seller notes meant for the shopper belong on the
 * customer timeline; the rest leak Stripe ids, webhook payloads or raw English
 * status echoes that the StatusPill already conveys.
 */
const INTERNAL_NOTE_PATTERNS: RegExp[] = [
	/\bintent=/i,
	/\bpi_[a-z0-9]/i,
	/\btx=/i,
	/webhook/i,
	/provider_(transaction|tx)_id/i,
	/^order created\b/i,
	/^payment (confirmed|captured|intent|succeeded|webhook)/i,
	/^auto-?cancelled\b/i,
	/reservation expired/i,
	/^cancelled by customer$/i,
	/seller fulfilled|all seller items fulfilled/i,
	/^tracking updated:/i,
	/^return requested\b/i,
];

export function customerFacingNote(note?: string | null): string | null {
	const trimmed = note?.trim();
	if (!trimmed) return null;
	return INTERNAL_NOTE_PATTERNS.some((pattern) => pattern.test(trimmed))
		? null
		: trimmed;
}
