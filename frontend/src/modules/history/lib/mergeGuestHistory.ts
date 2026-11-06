import { historyService } from "@/modules/history/api/historyService";
import {
	clearRecentlyViewed,
	getRecentlyViewedSlugs,
} from "@/shared/lib/recentlyViewed";

/**
 * Merges the guest cache into the authenticated account once per sign-in.
 * The cache is cleared only after the API accepts it, so a failed merge can
 * retry on the next authenticated boot.
 */
export async function mergeGuestHistory(): Promise<number> {
	const slugs = getRecentlyViewedSlugs();
	if (slugs.length === 0) return 0;

	try {
		const response = await historyService.syncGuest(slugs);
		clearRecentlyViewed();
		return response.data?.synced ?? 0;
	} catch {
		return 0;
	}
}
