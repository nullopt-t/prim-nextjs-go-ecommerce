import { describe, expect, it } from "vitest";
import { hasRoleRank, ROLE_RANKS } from "./roles";

describe("hasRoleRank", () => {
	it("lets a role meet its own tier and above", () => {
		expect(hasRoleRank("customer", ROLE_RANKS.customer)).toBe(true);
		expect(hasRoleRank("seller", ROLE_RANKS.seller)).toBe(true);
		expect(hasRoleRank("superadmin", ROLE_RANKS.admin)).toBe(true);
	});

	it("blocks lower tiers from higher consoles", () => {
		expect(hasRoleRank("customer", ROLE_RANKS.seller)).toBe(false);
		expect(hasRoleRank("seller", ROLE_RANKS.admin)).toBe(false);
	});

	it("treats missing/unknown roles as the lowest tier", () => {
		expect(hasRoleRank(undefined, 0)).toBe(true);
		expect(hasRoleRank(undefined, ROLE_RANKS.seller)).toBe(false);
		expect(hasRoleRank("bogus", ROLE_RANKS.seller)).toBe(false);
	});
});
