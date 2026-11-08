"use client";

import { Plus, Trash2 } from "lucide-react";
import { useTranslations } from "next-intl";
import type { SellerProduct } from "@/modules/seller/types";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/form/select";
import type { BundleItemDraft } from "./wizardTypes";

interface Props {
	products: SellerProduct[];
	items: BundleItemDraft[];
	currency: string;
	loading: boolean;
	onChange: (items: BundleItemDraft[]) => void;
	error?: React.ReactNode;
}

// Product-level composition only: a kit line bills whole units of a listed
// product, so the wizard leaves variant-level selection to the bundles page.
export function BundleItems({ products, items, currency, loading, onChange, error }: Props) {
	const t = useTranslations("seller");
	const totalMinor = items.reduce((sum, item) => {
		const unitPrice = products.find((product) => product.id === item.productId)?.price ?? 0;
		return sum + unitPrice * (Number.parseInt(item.quantity, 10) || 0);
	}, 0);
	const update = (index: number, patch: Partial<BundleItemDraft>) =>
		onChange(items.map((item, i) => (i === index ? { ...item, ...patch } : item)));

	return (
		<div className="space-y-3 rounded-md border border-border p-4">
			<div className="flex flex-wrap items-center justify-between gap-2">
				<div>
					<h3 className="text-sm font-semibold text-foreground">{t("bundles.itemsTitle")}</h3>
					<p className="text-xs text-muted-foreground">{t("wizard.bundle.subtitle")}</p>
				</div>
				<button
					type="button"
					onClick={() => onChange([...items, { productId: "", quantity: "1" }])}
					className="inline-flex items-center gap-1.5 rounded-md border border-border px-2.5 py-1.5 text-xs font-medium text-foreground hover:bg-secondary"
				>
					<Plus className="size-3.5" />
					{t("bundles.addItem")}
				</button>
			</div>

			{loading && <p className="text-xs text-muted-foreground">{t("wizard.bundle.loading")}</p>}
			{!loading && products.length === 0 && (
				<p className="rounded-md bg-secondary/50 p-3 text-xs text-muted-foreground">{t("wizard.bundle.noProducts")}</p>
			)}

			{items.map((item, index) => (
				<div key={index} className="flex items-end gap-2">
					<div className="min-w-0 flex-1 space-y-1">
						<label className="text-xs font-medium text-muted-foreground">{t("bundles.product")}</label>
						<Select value={item.productId} onValueChange={(value) => update(index, { productId: value })}>
							<SelectTrigger aria-label={`${t("bundles.product")} ${index + 1}`}>
								<SelectValue placeholder={t("bundles.product")} />
							</SelectTrigger>
							<SelectContent>
								{products.map((product) => (
									<SelectItem key={product.id} value={product.id}>
										{product.title}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</div>
					<div className="w-24 space-y-1">
						<label className="text-xs font-medium text-muted-foreground">{t("bundles.quantity")}</label>
						<input
							type="number"
							min="1"
							step="1"
							value={item.quantity}
							onChange={(e) => update(index, { quantity: e.target.value })}
							className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
						/>
					</div>
					<button
						type="button"
						disabled={items.length <= 1}
						onClick={() => onChange(items.filter((_, i) => i !== index))}
						className="rounded-md border border-border p-2 text-muted-foreground hover:bg-destructive/10 hover:text-destructive disabled:opacity-50"
						aria-label={t("actions.delete")}
					>
						<Trash2 className="size-4" />
					</button>
				</div>
			))}

			<div className="flex flex-wrap items-center gap-x-2 rounded-md bg-secondary/40 px-3 py-2 text-sm">
				<span className="text-muted-foreground">{t("bundles.itemsTotal")}</span>
				<span className="font-semibold tabular-nums text-foreground">
					{currency.toUpperCase() || "EGP"} {(totalMinor / 100).toFixed(2)}
				</span>
			</div>
			{error}
		</div>
	);
}
