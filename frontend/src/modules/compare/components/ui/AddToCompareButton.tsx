"use client";

import { useTranslations } from "next-intl";
import { Scale } from "lucide-react";
import { Link } from "@/shared/i18n/navigation";
import { useCompare } from "@/modules/compare/model/useCompare";
import { notify } from "@/shared/lib/notify";

interface AddToCompareButtonProps {
	slug?: string | null;
	productName?: string;
	className?: string;
}

/**
 * Totally optional "add to compare" trigger for PDPs. Self-contained: it reads
 * its own state from the compare context so it can drop into any leaf
 * ("use client") spot without plumbing props through presentational parents.
 */
export function AddToCompareButton({
	slug,
	productName,
	className = "",
}: AddToCompareButtonProps) {
	const t = useTranslations("compare");
	const { isCompared, toggleCompare, compareCount, canAdd } = useCompare();

	if (!slug) return null;

	const compared = isCompared(slug);

	const handleClick = () => {
		if (!slug) return;
		if (!compared && !canAdd) {
			notify.error(t("toast.limitReached"));
			return;
		}
		const ok = toggleCompare(slug);
		if (ok) {
			notify.success(
				compared
					? t("toast.removed")
					: t("toast.added", { name: productName || slug }),
			);
		} else if (!compared) {
			notify.error(t("toast.limitReached"));
		}
	};

	return (
		<div className={`flex items-center gap-3 ${className}`}>
			<button
				type="button"
				onClick={handleClick}
				aria-pressed={compared}
				className={`flex flex-1 cursor-pointer items-center justify-center gap-2 rounded-lg border px-4 py-2.5 text-xs font-medium transition-colors ${
					compared
						? "border-foreground/30 bg-secondary text-foreground"
						: "border-border bg-background text-muted-foreground hover:border-foreground/30 hover:text-foreground"
				}`}
			>
				<Scale className="size-3.5" aria-hidden="true" />
				{compared ? t("addedToCompare") : t("addToCompare")}
			</button>
			{compareCount > 0 && (
				<Link
					href="/compare"
					className="shrink-0 rounded-lg px-3 py-2.5 text-xs font-medium text-accent-brand hover:underline"
				>
					{t("viewCompare", { count: compareCount })}
				</Link>
			)}
		</div>
	);
}
