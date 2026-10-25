"use client";

import { useMemo } from "react";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
	Form,
	FormControl,
	FormField,
	FormInput,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/shared/ui/form/form";

export interface VariantBasicsValues {
	title: string;
	sku: string;
	price: string;
	originalPrice: string;
	stock: string;
	adjustmentReason: string;
}

interface VariantBasicsFieldsProps {
	value: VariantBasicsValues;
	mode: "create" | "edit";
	onChange: (patch: Partial<VariantBasicsValues>) => void;
}

export function VariantBasicsFields({
	value,
	mode,
	onChange,
}: VariantBasicsFieldsProps) {
	const t = useTranslations("admin");
	const schema = useMemo(
		() =>
			z.object({
				title: z
					.string()
					.trim()
					.min(1, t("products.variants.validation.titleRequired")),
				sku: z.string(),
				price: z
					.string()
					.min(1, t("products.variants.validation.priceRequired"))
					.refine(
						(input) =>
							Number.isFinite(Number.parseFloat(input)) &&
							Number.parseFloat(input) >= 0,
						t("products.variants.validation.priceInvalid"),
					),
				originalPrice: z.string().refine(
					(input) =>
						input.trim() === "" ||
						(Number.isFinite(Number.parseFloat(input)) &&
							Number.parseFloat(input) >= 0),
					t("products.variants.validation.originalPriceInvalid"),
				),
				stock: z
					.string()
					.min(1, t("products.variants.validation.stockRequired"))
					.refine(
						(input) => Number.isInteger(Number(input)) && Number(input) >= 0,
						t("products.variants.validation.stockInvalid"),
					),
				adjustmentReason: z.string(),
			}),
		[t],
	);

	const form = useForm<VariantBasicsValues>({
		resolver: zodResolver(schema),
		defaultValues: value,
		mode: "onTouched",
	});

	const inputClass =
		"bg-background px-3 py-2 focus:border-accent-brand focus:ring-accent-brand/30";
	const sync = (
		field: keyof VariantBasicsValues,
		onFieldChange: (next: string) => void,
	) => (event: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
		onFieldChange(event.target.value);
		onChange({ [field]: event.target.value });
	};

	return (
		<Form {...form}>
			<div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3">
				<FormField
					control={form.control}
					name="title"
					render={({ field }) => (
						<FormItem>
							<FormLabel className="text-[11px] font-semibold">
								{t("products.variants.fields.title")}
							</FormLabel>
							<FormControl>
								<FormInput
									{...field}
									className={inputClass}
									placeholder={t(
										"products.variants.fields.titlePlaceholder",
									)}
									onChange={sync("title", field.onChange)}
								/>
							</FormControl>
							<FormMessage />
						</FormItem>
					)}
				/>

				<FormField
					control={form.control}
					name="sku"
					render={({ field }) => (
						<FormItem>
							<FormLabel className="text-[11px] font-semibold">
								{t("products.variants.fields.sku")}
							</FormLabel>
							<FormControl>
								<FormInput
									{...field}
									className={`${inputClass} font-mono`}
									onChange={sync("sku", field.onChange)}
								/>
							</FormControl>
							<p className="text-[10px] text-muted-foreground">
								{t("products.form.skuHint")}
							</p>
							<FormMessage />
						</FormItem>
					)}
				/>

				<FormField
					control={form.control}
					name="price"
					render={({ field }) => (
						<FormItem>
							<FormLabel className="text-[11px] font-semibold">
								{t("products.variants.fields.price")}
							</FormLabel>
							<FormControl>
								<FormInput
									{...field}
									type="number"
									step="0.01"
									className={inputClass}
									onChange={sync("price", field.onChange)}
								/>
							</FormControl>
							<FormMessage />
						</FormItem>
					)}
				/>

				<FormField
					control={form.control}
					name="originalPrice"
					render={({ field }) => (
						<FormItem>
							<FormLabel className="text-[11px] font-semibold">
								{t("products.variants.fields.originalPrice")}
							</FormLabel>
							<FormControl>
								<FormInput
									{...field}
									type="number"
									step="0.01"
									className={inputClass}
									onChange={sync("originalPrice", field.onChange)}
								/>
							</FormControl>
							<FormMessage />
						</FormItem>
					)}
				/>

				<FormField
					control={form.control}
					name="stock"
					render={({ field }) => (
						<FormItem>
							<FormLabel className="text-[11px] font-semibold">
								{mode === "create"
									? t("products.variants.fields.initialStock")
									: t("products.form.stockQty")}{" "}
								*
							</FormLabel>
							<FormControl>
								<FormInput
									{...field}
									type="number"
									min="0"
									className={inputClass}
									onChange={sync("stock", field.onChange)}
								/>
							</FormControl>
							<FormMessage />
						</FormItem>
					)}
				/>

				{mode === "edit" && (
					<FormField
						control={form.control}
						name="adjustmentReason"
						render={({ field }) => (
							<FormItem>
								<FormLabel className="text-[11px] font-semibold">
									{t("products.form.auditReason")}
								</FormLabel>
								<FormControl>
									<select
										{...field}
										className="w-full rounded-xl border border-border bg-background px-3 py-2 text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
										onChange={sync("adjustmentReason", field.onChange)}
									>
										<option value="restock">
											{t("products.form.reasonRestock")}
										</option>
										<option value="adjustment">
											{t("products.form.reasonAdjustment")}
										</option>
										<option value="return">
											{t("products.form.reasonReturn")}
										</option>
										<option value="sale">
											{t("products.form.reasonSale")}
										</option>
									</select>
								</FormControl>
								<FormMessage />
							</FormItem>
						)}
					/>
				)}
			</div>
		</Form>
	);
}
