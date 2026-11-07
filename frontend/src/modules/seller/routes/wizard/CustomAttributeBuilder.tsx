"use client";

import { useState } from "react";
import { Plus } from "lucide-react";
import { sellerService } from "@/modules/seller/api/sellerService";
import type { SellerCustomAttribute } from "@/modules/seller/types";
import type { CategoryAttribute } from "@/modules/catalog/types";
import { apiErrorMessage, notify } from "@/shared/lib/notify";
import { useTranslations } from "next-intl";

interface Props {
	categoryId: string;
	onCreated: (attribute: CategoryAttribute) => void;
}

// Lets the seller define a custom attribute for the selected category; the
// new attribute is appended to the wizard's picker immediately.
export function CustomAttributeBuilder({ categoryId, onCreated }: Props) {
	const t = useTranslations("seller");
	const [key, setKey] = useState("");
	const [label, setLabel] = useState("");
	const [values, setValues] = useState("");
	const [creating, setCreating] = useState(false);

	const create = async () => {
		if (!categoryId || creating) return;
		const normalizedKey = key.trim().toLowerCase();
		const trimmedLabel = label.trim();
		const allowedValues = values
			.split(",")
			.map((value) => value.trim())
			.filter(Boolean);
		if (!normalizedKey || !trimmedLabel || allowedValues.length === 0) return;
		setCreating(true);
		try {
			const created = await sellerService.createAttribute({
				categoryId,
				key: normalizedKey,
				label: trimmedLabel,
				valueType: "enum",
				allowedValues,
			});
			setKey("");
			setLabel("");
			setValues("");
			const attribute = created.data as SellerCustomAttribute;
			onCreated({
				key: attribute.key,
				label: attribute.label,
				valueType: attribute.valueType || "enum",
				allowedValues: Array.isArray(attribute.allowedValues) ? attribute.allowedValues : allowedValues,
				isFilterable: attribute.isFilterable,
				sortOrder: attribute.sortOrder ?? 0,
			});
			notify.success(t("wizard.customAttributes.created"));
		} catch (error) {
			notify.error(apiErrorMessage(error, t("products.saveFailed")));
		} finally {
			setCreating(false);
		}
	};

	return (
		<div className="rounded-lg border border-dashed border-border p-3">
			<p className="text-xs font-semibold text-foreground">{t("wizard.customAttributes.title")}</p>
			<p className="mt-0.5 text-xs text-muted-foreground">{t("wizard.customAttributes.subtitle")}</p>
			<div className="mt-2 grid grid-cols-1 gap-2 sm:grid-cols-3">
				<input
					value={key}
					onChange={(e) => setKey(e.target.value)}
					placeholder={t("wizard.customAttributes.keyPlaceholder")}
					dir="ltr"
					className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
				/>
				<input
					value={label}
					onChange={(e) => setLabel(e.target.value)}
					placeholder={t("wizard.customAttributes.labelPlaceholder")}
					className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
				/>
				<input
					value={values}
					onChange={(e) => setValues(e.target.value)}
					placeholder={t("wizard.customAttributes.valuesPlaceholder")}
					className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
				/>
			</div>
			<button
				type="button"
				disabled={creating || !key.trim() || !label.trim() || !values.trim()}
				onClick={() => void create()}
				className="mt-2 inline-flex items-center gap-1.5 rounded-md border border-border px-2.5 py-1.5 text-xs font-medium text-foreground disabled:opacity-50"
			>
				<Plus className="size-3.5" />
				{t("wizard.customAttributes.create")}
			</button>
		</div>
	);
}
