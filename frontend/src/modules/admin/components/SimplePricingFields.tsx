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
import { CustomSelect } from "@/shared/ui/customSelect";

export interface SimplePricingValues {
	sku: string;
	price: string;
	compareAtPrice: string;
	stock: string;
	adjustmentReason: string;
}

interface SimplePricingFieldsProps {
	value: SimplePricingValues;
	hasChildren: boolean;
	onChange: (patch: Partial<SimplePricingValues>) => void;
}

export function SimplePricingFields({
	value,
	hasChildren,
	onChange,
}: SimplePricingFieldsProps) {
	const t = useTranslations("admin");
	const schema = useMemo(
		() =>
			z.object({
				sku: z.string(),
				price: z
					.string()
					.trim()
					.min(1, t("products.validation.priceRequired"))
					.refine(
						(v) => !Number.isNaN(Number(v)) && Number(v) >= 0,
						t("products.validation.priceInvalid"),
					),
				compareAtPrice: z.string().refine(
					(v) =>
						v.trim() === "" ||
						(!Number.isNaN(Number(v)) && Number(v) >= 0),
					t("products.validation.compareInvalid"),
				),
				stock: z
					.string()
					.trim()
					.min(1, t("products.validation.stockRequired"))
					.refine(
						(v) =>
							Number.isInteger(Number(v)) && Number(v) >= 0,
						t("products.validation.stockInvalid"),
					),
				adjustmentReason: z.string(),
			}),
		[t],
	);

	const form = useForm<SimplePricingValues>({
		resolver: zodResolver(schema),
		defaultValues: value,
		mode: "onTouched",
	});

	const inputClass =
		"bg-secondary/30 font-sans focus:border-accent-brand focus:ring-accent-brand/30";

	return (
		<Form {...form}>
			<div className="space-y-4 max-w-2xl">
				<div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
					<FormField
						control={form.control}
						name="sku"
						render={({ field }) => (
							<FormItem>
								<FormLabel className="font-bold">
									{t("products.form.sku")}
								</FormLabel>
								<FormControl>
									<FormInput
										{...field}
										className={`${inputClass} font-mono`}
										onChange={(e) => {
											field.onChange(e);
											onChange({ sku: e.target.value });
										}}
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
								<FormLabel className="font-bold">
									{t("products.form.price")} *
								</FormLabel>
								<FormControl>
									<FormInput
										{...field}
										type="number"
										step="0.01"
										className={inputClass}
										onChange={(e) => {
											field.onChange(e);
											onChange({ price: e.target.value });
										}}
									/>
								</FormControl>
								<FormMessage />
							</FormItem>
						)}
					/>
					<FormField
						control={form.control}
						name="compareAtPrice"
						render={({ field }) => (
							<FormItem>
								<FormLabel className="font-bold">
									{t("products.form.comparePrice")}
								</FormLabel>
								<FormControl>
									<FormInput
										{...field}
										type="number"
										step="0.01"
										className={inputClass}
										onChange={(e) => {
											field.onChange(e);
											onChange({ compareAtPrice: e.target.value });
										}}
									/>
								</FormControl>
								<p className="text-[10px] text-muted-foreground">
									{t("products.form.comparePriceHint")}
								</p>
								<FormMessage />
							</FormItem>
						)}
					/>
					<FormField
						control={form.control}
						name="stock"
						render={({ field }) => (
							<FormItem>
								<FormLabel className="font-bold">
									{t("products.form.stockQty")} *
								</FormLabel>
								<FormControl>
									<FormInput
										{...field}
										type="number"
										min="0"
										className={inputClass}
										onChange={(e) => {
											field.onChange(e);
											onChange({ stock: e.target.value });
										}}
									/>
								</FormControl>
								<p className="text-[10px] text-muted-foreground">
									{t("products.form.stockHint")}
								</p>
								<FormMessage />
							</FormItem>
						)}
					/>
				</div>

				{hasChildren && (
					<div className="space-y-1.5 pt-3 border-t border-border">
						<FormField
							control={form.control}
							name="adjustmentReason"
							render={({ field }) => (
								<FormItem>
									<FormLabel className="font-bold">
										{t("products.form.auditReason")}
									</FormLabel>
									<FormControl>
										<CustomSelect
											name={field.name}
											ref={field.ref}
											value={field.value ?? ""}
											onChange={(value) => {
												field.onChange(value);
												onChange({ adjustmentReason: value });
											}}
											options={[
												{ value: "adjustment", label: t("products.form.reasonAdjustment") },
												{ value: "restock", label: t("products.form.reasonRestock") },
												{ value: "return", label: t("products.form.reasonReturn") },
												{ value: "sale", label: t("products.form.reasonSale") },
											]}
											size="sm"
											className="w-full"
										/>
									</FormControl>
									<p className="text-[11px] text-muted-foreground">
										{t("products.form.auditNote")}
									</p>
									<FormMessage />
								</FormItem>
							)}
						/>
					</div>
				)}
			</div>
		</Form>
	);
}
