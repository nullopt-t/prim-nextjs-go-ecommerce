"use client";

import { Loader2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { combinationKey, formatCombination, type VariantOverride } from "./wizardTypes";

interface Props {
	currency: string;
	onCurrencyChange: (value: string) => void;
	currencyError: React.ReactNode;
	combinations: Record<string, string>[];
	variantOverrides: Record<string, VariantOverride>;
	setVariantOverrides: React.Dispatch<React.SetStateAction<Record<string, VariantOverride>>>;
	uploadingThumbnailKey: string | null;
	onUploadThumbnail: (key: string, file: File | undefined) => void;
}

// Per-variant pricing, SKU, stock, and thumbnail editor for variable
// products. Currency is shared; everything else is per combination.
export function VariantPricing({
	currency,
	onCurrencyChange,
	currencyError,
	combinations,
	variantOverrides,
	setVariantOverrides,
	uploadingThumbnailKey,
	onUploadThumbnail,
}: Props) {
	const t = useTranslations("seller");

	return (
		<>
			<div className="max-w-40 space-y-1">
				<label htmlFor="wizard-currency" className="text-sm font-medium text-foreground">{t("products.form.currency")}</label>
				<input
					id="wizard-currency"
					value={currency}
					onChange={(e) => onCurrencyChange(e.target.value.toUpperCase())}
					className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
				/>
				{currencyError}
			</div>
			{combinations.length > 0 && (
				<div className="space-y-3 rounded-md border border-border p-4">
					<div>
						<h3 className="text-sm font-semibold text-foreground">{t("wizard.variants.title")}</h3>
						<p className="text-xs text-muted-foreground">{t("wizard.variants.subtitle")}</p>
					</div>
					{combinations.map((combination) => {
						const key = combinationKey(combination);
						const override = variantOverrides[key] ?? {};
						const setOverride = (patch: VariantOverride) =>
							setVariantOverrides((current) => ({ ...current, [key]: { ...(current[key] ?? {}), ...patch } }));
						return (
							<div key={key} className="rounded-lg border border-border p-3">
								<p className="text-sm font-medium text-foreground">{formatCombination(combination)}</p>
								<div className="mt-2 grid grid-cols-2 gap-3 sm:grid-cols-4">
									<div className="space-y-1">
										<label className="text-xs font-medium text-muted-foreground">{t("products.form.sellingPrice")}</label>
										<input
											type="number"
											min="0.01"
											step="0.01"
											value={override.price ?? ""}
											onChange={(e) => setOverride({ price: e.target.value })}
											className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
										/>
									</div>
									<div className="space-y-1">
										<label className="text-xs font-medium text-muted-foreground">{t("products.form.compareAtPrice")}</label>
										<input
											type="number"
											min="0"
											step="0.01"
											value={override.crossesOutPrice ?? ""}
											onChange={(e) => setOverride({ crossesOutPrice: e.target.value })}
											className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
										/>
									</div>
									<div className="space-y-1">
										<label className="text-xs font-medium text-muted-foreground">{t("products.form.skuCode")}</label>
										<input
											value={override.sku ?? ""}
											onChange={(e) => setOverride({ sku: e.target.value })}
											className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
										/>
									</div>
									<div className="space-y-1">
										<label className="text-xs font-medium text-muted-foreground">{t("products.form.initialStock")}</label>
										<input
											type="number"
											min="0"
											value={override.stock ?? ""}
											onChange={(e) => setOverride({ stock: e.target.value })}
											className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
										/>
									</div>
								</div>
								<div className="mt-3 flex items-center gap-3">
									{override.thumbnailUrl ? (
										<img src={override.thumbnailUrl} alt={formatCombination(combination)} className="size-14 rounded-md border border-border object-cover" />
									) : (
										<div className="flex size-14 items-center justify-center rounded-md border border-dashed border-border text-xs text-muted-foreground">
											{t("wizard.variants.thumbnail")}
										</div>
									)}
									<div className="flex flex-wrap items-center gap-2">
										<label className="inline-flex cursor-pointer items-center rounded-md border border-border px-2.5 py-1.5 text-xs font-medium text-foreground hover:bg-secondary">
											{uploadingThumbnailKey === key ? <Loader2 className="size-3.5 animate-spin" /> : null}
											{t("wizard.variants.chooseThumbnail")}
											<input
												type="file"
												accept="image/*"
												disabled={uploadingThumbnailKey === key}
												className="sr-only"
												onChange={(e) => {
													onUploadThumbnail(key, e.target.files?.[0]);
													e.target.value = "";
												}}
											/>
										</label>
										{override.thumbnailUrl && (
											<button
												type="button"
												onClick={() => setOverride({ thumbnailUrl: undefined })}
												className="rounded-md border border-border px-2.5 py-1.5 text-xs font-medium text-muted-foreground hover:text-destructive"
											>
												{t("wizard.variants.removeThumbnail")}
											</button>
										)}
									</div>
								</div>
							</div>
						);
					})}
				</div>
			)}
		</>
	);
}
