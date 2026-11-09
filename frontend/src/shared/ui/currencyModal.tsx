"use client";

import { Check, Coins } from "lucide-react";
import { useTranslations } from "next-intl";
import {
	useDisplayCurrency,
	usePublicSettings,
} from "@/modules/settings/model/publicSettingsContext";
import {
	SUPPORTED_CURRENCIES,
	CURRENCY_METADATA,
	type SupportedCurrency,
} from "@/shared/lib/currency";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogHeader,
	DialogTitle,
} from "@/shared/ui/form/dialog";

interface CurrencyModalProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

export function CurrencyModal({ open, onOpenChange }: CurrencyModalProps) {
	const t = useTranslations("common.currency");
	const displayCurrency = useDisplayCurrency();
	const { setDisplayCurrency } = usePublicSettings();

	const handleSelect = (code: SupportedCurrency) => {
		setDisplayCurrency(code);
		onOpenChange(false);
	};

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="max-w-md p-6">
				<DialogHeader>
					<div className="flex items-center gap-2.5">
						<div className="flex size-9 items-center justify-center rounded-xl bg-accent-brand/10 text-accent-brand">
							<Coins className="size-5" aria-hidden="true" />
						</div>
						<DialogTitle className="text-lg font-bold text-foreground">
							{t("modalTitle")}
						</DialogTitle>
					</div>
					<DialogDescription className="mt-1.5 text-xs text-muted-foreground leading-relaxed">
						{t("modalSubtitle")}
					</DialogDescription>
				</DialogHeader>

				<div
					role="radiogroup"
					aria-label={t("label")}
					className="mt-5 flex flex-col gap-2.5"
				>
					{SUPPORTED_CURRENCIES.map((code) => {
						const isSelected = code === displayCurrency;
						const meta = CURRENCY_METADATA[code];
						const name = t(code);

						return (
							<button
								key={code}
								type="button"
								role="radio"
								aria-checked={isSelected}
								onClick={() => handleSelect(code)}
								className={`group flex items-center justify-between p-3.5 rounded-2xl border transition-all text-start cursor-pointer ${
									isSelected
										? "border-accent-brand bg-accent-brand/5 shadow-xs ring-1 ring-accent-brand"
										: "border-border/70 hover:border-accent-brand/40 hover:bg-secondary/60 bg-card"
								}`}
							>
								<div className="flex items-center gap-3.5 min-w-0">
									<span
										className="text-2xl leading-none select-none shrink-0"
										role="img"
										aria-label={meta.country}
									>
										{meta.flag}
									</span>
									<div className="flex flex-col min-w-0">
										<div className="flex items-center gap-2">
											<span className="font-bold text-sm text-foreground">
												{code}
											</span>
											<span className="text-xs text-muted-foreground font-mono font-medium">
												({meta.symbol})
											</span>
										</div>
										<span className="text-xs text-muted-foreground truncate">
											{name}
										</span>
									</div>
								</div>

								{isSelected && (
									<div className="size-6 rounded-full bg-accent-brand text-white flex items-center justify-center shrink-0 shadow-2xs">
										<Check className="size-3.5 stroke-[2.5]" aria-hidden="true" />
									</div>
								)}
							</button>
						);
					})}
				</div>

				<div className="mt-5 pt-4 border-t border-border/60">
					<p className="text-[11px] text-muted-foreground text-center">
						{t("approximate")}
					</p>
				</div>
			</DialogContent>
		</Dialog>
	);
}
