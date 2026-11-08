import { historyService } from "@/modules/history/api/historyService";
import {
	getRecentlyViewedSlugs,
	removeRecentlyViewed,
} from "@/shared/lib/recentlyViewed";

// One in-flight sync per page load. fetchUser runs on mount and again after
// verify, and a second caller would resend the same slugs and race the first
// one's cleanup.
let inFlight: Promise<number> | null = null;

async function runMerge(): Promise<number> {
	const slugs = getRecentlyViewedSlugs();
	if (slugs.length === 0) return 0;

	try {
		const response = await historyService.syncGuest(slugs);
		// Remove only the slugs that were synced. Views recorded while the
		// sync was in flight stay in the guest cache for the next merge
		// instead of being wiped by a blanket clear.
		for (const slug of slugs) removeRecentlyViewed(slug);
		return response.data?.synced ?? 0;
	} catch {
		return 0;
	}
}

/**
 * Merges the guest cache into the authenticated account once per sign-in.
 * The cache is cleared only after the API accepts it, so a failed merge can
 * retry on the next authenticated boot.
 */
export function mergeGuestHistory(): Promise<number> {
	if (!inFlight) {
		inFlight = runMerge().finally(() => {
			inFlight = null;
		});
	}
	return inFlight;
}
