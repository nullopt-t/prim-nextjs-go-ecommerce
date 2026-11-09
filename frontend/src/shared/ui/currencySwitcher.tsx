"use client";

import { useState } from "react";
import { ChevronDown, Coins } from "lucide-react";
import { useTranslations } from "next-intl";
import { useDisplayCurrency } from "@/modules/settings/model/publicSettingsContext";
import { getCurrencyMeta } from "@/shared/lib/currency";
import { CurrencyModal } from "@/shared/ui/currencyModal";

interface CurrencySwitcherProps {
	/** "menu-item" = row inside header dropdown; "footer" = footer trigger; "header" = compact icon button. */
	variant?: "header" | "footer" | "menu-item";
	/** Optional callback when clicked (e.g. to close parent dropdown). */
	onClick?: () => void;
	/** Optional controlled open state for the currency modal. */
	open?: boolean;
	/** Optional controlled open change callback. */
	onOpenChange?: (open: boolean) => void;
}

/**
 * Currency switcher trigger (dropdown menu item, footer, or header).
 * Clicking opens a popup modal with the list of currencies and their flags.
 */
export function CurrencySwitcher({
	variant = "menu-item",
	onClick,
	open: controlledOpen,
	onOpenChange: controlledOnOpenChange,
}: CurrencySwitcherProps) {
	const t = useTranslations("common.currency");
	const displayCurrency = useDisplayCurrency();
	const [internalOpen, setInternalOpen] = useState(false);

	const isControlled = controlledOpen !== undefined;
	const isModalOpen = isControlled ? controlledOpen : internalOpen;
	const setModalOpen = isControlled ? (controlledOnOpenChange ?? (() => {})) : setInternalOpen;

	const meta = getCurrencyMeta(displayCurrency);

	const handleClick = () => {
		if (onClick) {
			onClick();
		} else {
			setModalOpen(true);
		}
	};

	if (variant === "menu-item") {
		return (
			<>
				<button
					type="button"
					onClick={handleClick}
					className="w-full flex items-center justify-between px-3 py-2 rounded-xl text-xs font-medium text-foreground hover:bg-secondary/70 transition-colors cursor-pointer"
				>
					<div className="flex items-center gap-2.5">
						<Coins className="size-4 text-muted-foreground" aria-hidden="true" />
						<span>{t("label")}</span>
					</div>
					<div className="flex items-center gap-1.5 text-[11px] font-semibold bg-secondary/80 px-2 py-0.5 rounded-md border border-border/50">
						<span className="text-sm leading-none" role="img" aria-label={meta.country}>
							{meta.flag}
						</span>
						<span className="font-mono">{displayCurrency || "EGP"}</span>
					</div>
				</button>

				{!isControlled && (
					<CurrencyModal open={isModalOpen} onOpenChange={setModalOpen} />
				)}
			</>
		);
	}

	if (variant === "footer") {
		return (
			<>
				<button
					type="button"
					onClick={handleClick}
					aria-label={t("label")}
					className="inline-flex items-center gap-2 rounded-lg border border-border bg-background px-3 py-2 text-xs font-semibold text-foreground hover:border-accent-brand/50 cursor-pointer shadow-2xs"
				>
					<span className="text-base leading-none" role="img" aria-label={meta.country}>
						{meta.flag}
					</span>
					<span className="font-mono">{displayCurrency || "—"}</span>
					<ChevronDown className="size-3.5 text-muted-foreground" aria-hidden="true" />
				</button>

				{!isControlled && (
					<CurrencyModal open={isModalOpen} onOpenChange={setModalOpen} />
				)}
			</>
		);
	}

	// variant === "header" (compact icon button fallback)
	return (
		<>
			<button
				type="button"
				onClick={handleClick}
				aria-label={t("label")}
				className="size-9 sm:h-10 sm:w-auto rounded-xl text-muted-foreground hover:text-accent-brand hover:bg-secondary/70 transition-all flex items-center gap-1.5 border border-border/40 hover:border-accent-brand/40 shadow-2xs p-0 sm:px-2.5 justify-center sm:justify-start text-xs font-semibold cursor-pointer shrink-0"
			>
				<span className="text-sm leading-none" role="img" aria-label={meta.country}>
					{meta.flag}
				</span>
				<span className="hidden sm:inline font-mono">{displayCurrency || "—"}</span>
				<ChevronDown className="size-3.5 hidden sm:block text-muted-foreground" aria-hidden="true" />
			</button>

			{!isControlled && (
				<CurrencyModal open={isModalOpen} onOpenChange={setModalOpen} />
			)}
		</>
	);
}
