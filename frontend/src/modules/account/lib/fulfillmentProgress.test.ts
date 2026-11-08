import { describe, it, expect } from "vitest";
import {
	MILESTONES,
	deriveMilestoneKey,
	resolveMilestone,
	stepState,
} from "./fulfillmentProgress";

describe("fulfillmentProgress (docs/order-milestones.md contract)", () => {
	it("backend substatus + percent win over status derivation", () => {
		const r = resolveMilestone("shipped", "out_for_delivery", 50, false);
		expect(r).toEqual({ index: 4, percent: 50 });
	});

	it("falls back to status derivation when backend fields are absent", () => {
		expect(resolveMilestone("pending")).toEqual({ index: 0, percent: 0 });
		expect(resolveMilestone("processing")).toEqual({ index: 1, percent: 5 });
		expect(resolveMilestone("shipped", null, null, false)).toEqual({ index: 2, percent: 10 });
		expect(resolveMilestone("shipped", null, null, true)).toEqual({ index: 4, percent: 50 });
		expect(resolveMilestone("delivered")).toEqual({ index: 5, percent: 100 });
	});

	it("ignores unknown substatus and clamps bad percents", () => {
		expect(resolveMilestone("processing", "bogus", 999)).toEqual({ index: 1, percent: 100 });
		expect(resolveMilestone("processing", null, -40)).toEqual({ index: 1, percent: 0 });
		expect(resolveMilestone("processing", null, NaN)).toEqual({ index: 1, percent: 5 });
	});

	it("deriveMilestoneKey mirrors the backend derivation", () => {
		expect(deriveMilestoneKey("pending", false)).toBe("pre_transit");
		expect(deriveMilestoneKey("cancelled", false)).toBe("pre_transit");
		expect(deriveMilestoneKey("shipped", true)).toBe("out_for_delivery");
	});

	it("milestone percents are monotonic and end at 100", () => {
		const percents = MILESTONES.map((m) => m.percent);
		expect(percents).toEqual([0, 5, 10, 25, 50, 100]);
		for (let i = 1; i < percents.length; i++) {
			expect(percents[i]).toBeGreaterThan(percents[i - 1]);
		}
	});

	it("stepState marks done/current/upcoming around the current index", () => {
		expect(stepState(0, 2)).toBe("done");
		expect(stepState(2, 2)).toBe("current");
		expect(stepState(4, 2)).toBe("upcoming");
	});
});
