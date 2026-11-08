/**
 * Fulfillment milestone timeline (see docs/order-milestones.md).
 *
 * The backend is the source of truth: every order response carries
 * `substatus` + `milestonePercent`. The table below mirrors that contract as
 * a fallback for stale/cached payloads that predate the fields — the backend
 * value always wins when present.
 */
export interface Milestone {
	key: string;
	percent: number;
}

export const MILESTONES: Milestone[] = [
	{ key: "pre_transit", percent: 0 },
	{ key: "picked_at_warehouse", percent: 5 },
	{ key: "arrived_regional_hub", percent: 10 },
	{ key: "arrived_local_depot", percent: 25 },
	{ key: "out_for_delivery", percent: 50 },
	{ key: "delivered", percent: 100 },
];

const MILESTONE_INDEX: Record<string, number> = Object.fromEntries(
	MILESTONES.map((m, i) => [m.key, i]),
);

/** Status-only derivation fallback (mirrors backend DeriveMilestone). */
export function deriveMilestoneKey(status: string, hasTracking: boolean): string {
	switch ((status || "pending").toLowerCase()) {
		case "processing":
			return "picked_at_warehouse";
		case "shipped":
			return hasTracking ? "out_for_delivery" : "arrived_regional_hub";
		case "delivered":
			return "delivered";
		default:
			return "pre_transit";
	}
}

export interface ResolvedMilestone {
	/** 0-based index into MILESTONES. */
	index: number;
	/** Filled-track percent (backend milestonePercent wins when present). */
	percent: number;
}

/**
 * Resolves the current milestone for an order, preferring the backend
 * `substatus`/`milestonePercent` and clamping to the table so a bad payload
 * can never render a negative or >100 width.
 */
export function resolveMilestone(
	status: string,
	substatus?: string | null,
	milestonePercent?: number | null,
	hasTracking?: boolean,
): ResolvedMilestone {
	const key = substatus && MILESTONE_INDEX[substatus] !== undefined
		? substatus
		: deriveMilestoneKey(status, Boolean(hasTracking));
	const index = MILESTONE_INDEX[key] ?? 0;
	const fallback = MILESTONES[index].percent;
	const percent = typeof milestonePercent === "number" && !Number.isNaN(milestonePercent)
		? Math.min(100, Math.max(0, milestonePercent))
		: fallback;
	return { index, percent };
}

/** Stepper state for one milestone position relative to the current index. */
export function stepState(index: number, currentIndex: number): "done" | "current" | "upcoming" {
	if (index < currentIndex) return "done";
	if (index === currentIndex) return "current";
	return "upcoming";
}
