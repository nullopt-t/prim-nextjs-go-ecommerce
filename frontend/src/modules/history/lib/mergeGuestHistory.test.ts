import { beforeEach, describe, expect, it, vi } from "vitest";

const removed: string[] = [];

vi.mock("@/modules/history/api/historyService", () => ({
	historyService: {
		syncGuest: vi.fn(async (slugs: string[]) => ({
			data: { synced: slugs.length },
		})),
	},
}));

vi.mock("@/shared/lib/recentlyViewed", () => ({
	getRecentlyViewedSlugs: vi.fn(() => ["a", "b"]),
	removeRecentlyViewed: vi.fn((slug?: string | null) => {
		if (slug) removed.push(slug);
	}),
	clearRecentlyViewed: vi.fn(),
}));

import { historyService } from "@/modules/history/api/historyService";
import { mergeGuestHistory } from "./mergeGuestHistory";

describe("mergeGuestHistory", () => {
	beforeEach(() => {
		removed.length = 0;
		vi.clearAllMocks();
	});

	it("sends the guest slugs once and removes only those slugs", async () => {
		const synced = await mergeGuestHistory();

		expect(synced).toBe(2);
		expect(historyService.syncGuest).toHaveBeenCalledTimes(1);
		expect(historyService.syncGuest).toHaveBeenCalledWith(["a", "b"]);
		expect(removed).toEqual(["a", "b"]);
	});

	it("shares one in-flight sync between concurrent callers", async () => {
		const [first, second] = await Promise.all([
			mergeGuestHistory(),
			mergeGuestHistory(),
		]);

		expect(first).toBe(2);
		expect(second).toBe(2);
		expect(historyService.syncGuest).toHaveBeenCalledTimes(1);
		expect(removed).toEqual(["a", "b"]);
	});

	it("keeps the guest cache when the sync fails", async () => {
		vi.mocked(historyService.syncGuest).mockRejectedValueOnce(
			new Error("network"),
		);

		const synced = await mergeGuestHistory();

		expect(synced).toBe(0);
		expect(removed).toEqual([]);
	});
});
