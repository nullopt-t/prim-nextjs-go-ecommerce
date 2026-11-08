"use client";

import { useEffect, useRef, useState } from "react";
import { ChevronDown, Coins } from "lucide-react";
import { useTranslations } from "next-intl";
import {
	useDisplayCurrency,
	usePublicSettings,
} from "@/modules/settings/model/publicSettingsContext";
import { SUPPORTED_CURRENCIES } from "@/shared/lib/currency";

interface CurrencySwitcherProps {
	/** "header" = compact icon button; "footer" = label + list button. */
	variant?: "header" | "footer";
}

/**
 * Currency switcher (header + footer). Changing it re-renders all converted
 * prices from the public-settings context; the choice persists via cookie +
 * localStorage. Converted amounts are display estimates (see currency.ts).
 */
export function CurrencySwitcher({ variant = "header" }: CurrencySwitcherProps) {
	const t = useTranslations("common.currency");
	const displayCurrency = useDisplayCurrency();
	const { setDisplayCurrency } = usePublicSettings();
	const [open, setOpen] = useState(false);
	const ref = useRef<HTMLDivElement>(null);

	useEffect(() => {
		if (!open) return;
		const onClick = (event: MouseEvent) => {
			if (ref.current && !ref.current.contains(event.target as Node)) setOpen(false);
		};
		const onKey = (event: KeyboardEvent) => {
			if (event.key === "Escape") setOpen(false);
		};
		document.addEventListener("mousedown", onClick);
		document.addEventListener("keydown", onKey);
		return () => {
			document.removeEventListener("mousedown", onClick);
			document.removeEventListener("keydown", onKey);
		};
	}, [open]);

	const options = SUPPORTED_CURRENCIES.map((code) => ({
		code,
		name: t(code),
	}));

	return (
		<div className="relative" ref={ref}>
			<button
				type="button"
				onClick={() => setOpen((prev) => !prev)}
				aria-expanded={open}
				aria-haspopup="listbox"
				aria-label={t("label")}
				className={
					variant === "header"
						? "h-9 sm:h-10 rounded-xl text-muted-foreground hover:text-accent-brand hover:bg-secondary/70 transition-all flex items-center gap-1.5 border border-border/40 hover:border-accent-brand/40 shadow-2xs px-2 sm:px-2.5 text-xs font-semibold cursor-pointer"
						: "inline-flex items-center gap-2 rounded-lg border border-border bg-background px-3 py-2 text-xs font-semibold text-foreground hover:border-accent-brand/50 cursor-pointer"
				}
			>
				<Coins className="size-4" aria-hidden="true" />
				<span className={variant === "header" ? "hidden sm:inline" : undefined}>{displayCurrency || "—"}</span>
				<ChevronDown
					className={`size-3.5 transition-transform ${variant === "header" ? "hidden sm:block" : ""} ${open ? "rotate-180" : ""}`}
					aria-hidden="true"
				/>
			</button>

			{open && (
				<div
					role="listbox"
					aria-label={t("label")}
					className={`absolute z-50 w-56 rounded-2xl border border-border/80 bg-card py-2 shadow-xl animate-in fade-in-0 zoom-in-95 duration-150 ${
						variant === "header" ? "end-0 mt-2 origin-top-end" : "bottom-full mb-2 start-0 origin-bottom-start"
					}`}
				>
					{options.map(({ code, name }) => {
						const selected = code === displayCurrency;
						return (
							<button
								key={code}
								type="button"
								role="option"
								aria-selected={selected}
								onClick={() => {
									setDisplayCurrency(code);
									setOpen(false);
								}}
								className={`flex w-full items-center justify-between px-3 py-2 text-xs transition-colors cursor-pointer ${
									selected
										? "font-bold text-accent-brand"
										: "text-muted-foreground hover:text-foreground hover:bg-secondary/70"
								}`}
							>
								<span>{name}</span>
								<span className="font-mono font-semibold">{code}</span>
							</button>
						);
					})}
					<p className="border-t border-border/60 px-3 py-2 text-[11px] text-muted-foreground">
						{t("approximate")}
					</p>
				</div>
			)}
		</div>
	);
}
