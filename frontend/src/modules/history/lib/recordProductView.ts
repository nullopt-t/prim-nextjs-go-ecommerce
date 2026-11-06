import { historyService } from "@/modules/history/api/historyService";
import { RECENTLY_VIEWED_UPDATED_EVENT } from "@/shared/lib/recentlyViewed";

/**
 * Fire-and-forget authenticated view telemetry. The payload is small and
 * keepalive lets it complete if the shopper navigates immediately.
 */
export function recordProductView(productId?: string | null) {
	if (!productId || typeof window === "undefined") return;

	void historyService
		.recordView(productId)
		.then(() => {
			window.dispatchEvent(new Event(RECENTLY_VIEWED_UPDATED_EVENT));
		})
		.catch(() => {
			// Browsing history is best-effort; never interrupt shopping.
		});
}
