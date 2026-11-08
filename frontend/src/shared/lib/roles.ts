export const ROLE_RANKS: Record<string, number> = {
  customer: 1,
  seller: 2,
  support: 3,
  admin: 4,
  superadmin: 5,
};

/** UX-level rank check (e.g. hiding consoles a session cannot enter). The
 * backend remains the authorization authority; unknown roles rank lowest. */
export function hasRoleRank(role: string | undefined | null, minRank: number): boolean {
  return (ROLE_RANKS[role ?? "customer"] ?? 0) >= minRank;
}
