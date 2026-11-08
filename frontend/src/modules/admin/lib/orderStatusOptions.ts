/**
 * Allowed next order statuses for the admin status selector. Orders in any
 * other status (paid, refunded, …) must still render their current status as
 * a disabled option — otherwise the select's value matches no <option> and
 * the browser renders it blank.
 */
const TRANSITIONS: Record<string, string[]> = {
	pending: ["processing", "cancelled"],
	processing: ["shipped", "cancelled"],
	shipped: ["delivered", "cancelled"],
	delivered: [],
	cancelled: [],
};

export function orderStatusOptions(status: string): string[] {
	const current = (status || "pending").toLowerCase();
	return [current, ...(TRANSITIONS[current] ?? [])];
}
