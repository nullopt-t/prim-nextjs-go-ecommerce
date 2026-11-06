import type { ReactNode } from "react";

export type StatusTone = "neutral" | "success" | "warning" | "danger" | "info";

const TONE_CLASSES: Record<StatusTone, string> = {
	neutral:
		"border-border bg-secondary/60 text-secondary-foreground",
	success:
		"border-emerald-600/20 bg-emerald-600/10 text-emerald-700 dark:text-emerald-400",
	warning:
		"border-amber-600/20 bg-amber-600/10 text-amber-700 dark:text-amber-400",
	danger:
		"border-destructive/20 bg-destructive/10 text-destructive",
	info: "border-sky-600/20 bg-sky-600/10 text-sky-700 dark:text-sky-400",
};

/**
 * Small, restrained status label for tables and detail panels.
 *
 * Uses a bordered pill instead of saturated fills so status stays readable in
 * both light and dark themes without shouting.
 */
export function StatusPill({
	children,
	tone = "neutral",
}: {
	children: ReactNode;
	tone?: StatusTone;
}) {
	return (
		<span
			className={`inline-flex items-center whitespace-nowrap rounded-md border px-2 py-0.5 text-xs font-medium ${TONE_CLASSES[tone]}`}
		>
			{children}
		</span>
	);
}

export function statusTone(status: string | undefined | null): StatusTone {
	switch ((status ?? "").toLowerCase()) {
		case "active":
		case "published":
		case "paid":
		case "delivered":
		case "approved":
		case "in stock":
		case "healthy":
			return "success";
		case "pending":
		case "processing":
		case "requested":
		case "low stock":
		case "draft":
		case "degraded":
			return "warning";
		case "cancelled":
		case "canceled":
		case "rejected":
		case "suspended":
		case "out of stock":
		case "unhealthy":
			return "danger";
		case "shipped":
		case "returned":
			return "info";
		default:
			return "neutral";
	}
}
