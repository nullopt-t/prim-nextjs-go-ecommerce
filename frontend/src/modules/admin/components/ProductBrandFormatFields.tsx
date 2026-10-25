"use client";

import { useMemo } from "react";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Check } from "lucide-react";
import {
	Form,
	FormControl,
	FormField,
	FormItem,
	FormLabel,
} from "@/shared/ui/form/form";

export interface ProductBrandFormatValues {
	brandId: string;
	productType: "simple" | "variable";
}

interface ProductBrandFormatFieldsProps {
	value: ProductBrandFormatValues;
	brands: Array<{ id: string; name: string }>;
	currentCategoryName: string;
	axesLocked: boolean;
	onChange: (patch: {
		brandId?: string;
		brand?: string;
		productType?: "simple" | "variable";
		variationAxes?: string[];
	}) => void;
}

export function ProductBrandFormatFields({
	value,
	brands,
	currentCategoryName,
	axesLocked,
	onChange,
}: ProductBrandFormatFieldsProps) {
	const t = useTranslations("admin");
	const schema = useMemo(
		() =>
			z.object({
				brandId: z.string(),
				productType: z.enum(["simple", "variable"]),
			}),
		[],
	);

	const form = useForm<ProductBrandFormatValues>({
		resolver: zodResolver(schema),
		defaultValues: value,
	});

	const selectProductType = (productType: "simple" | "variable") => {
		if (axesLocked) return;
		form.setValue("productType", productType, { shouldValidate: true });
		onChange(
			productType === "simple"
				? { productType, variationAxes: [] }
				: { productType },
		);
	};

	return (
		<Form {...form}>
			<div className="space-y-6">
				<FormField
					control={form.control}
					name="brandId"
					render={({ field }) => (
						<FormItem className="space-y-2 rounded-xl border border-border bg-secondary/20 p-4 max-w-xl">
							<div className="flex items-center justify-between">
								<FormLabel className="font-bold">
									{t("products.form.brand")}
								</FormLabel>
								<span className="text-[11px] text-muted-foreground">
									{t("products.wizard.brandCategoryHint", {
										category: currentCategoryName,
									})}
								</span>
							</div>
							<FormControl>
								<select
									{...field}
									className="w-full px-3.5 py-2.5 bg-background border border-border rounded-xl text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
									onChange={(e) => {
										field.onChange(e);
										const brand = brands.find(
											(b) => b.id === e.target.value,
										);
										onChange({
											brandId: e.target.value,
											brand: brand?.name || "",
										});
									}}
								>
									<option value="">
										{t("products.form.noBrand")}
									</option>
									{brands.map((brand, index) => (
										<option
											key={`brand-opt-${brand.id || index}`}
											value={brand.id}
										>
											{brand.name}
										</option>
									))}
								</select>
							</FormControl>
						</FormItem>
					)}
				/>

				<FormField
					control={form.control}
					name="productType"
					render={({ field }) => (
						<FormItem className="space-y-3">
							<FormLabel className="font-bold">
								{t("products.form.productType")}
							</FormLabel>
							<div className="grid grid-cols-1 sm:grid-cols-2 gap-4 max-w-2xl">
								{(
									[
										{
											type: "simple" as const,
											label: t("products.form.typeSimple"),
											description: t(
												"products.wizard.simpleProductDesc",
											),
										},
										{
											type: "variable" as const,
											label: t("products.form.typeVariable"),
											description: t(
												"products.wizard.variableProductDesc",
											),
										},
									] as const
								).map((option) => {
									const active = field.value === option.type;
									return (
										<button
											key={option.type}
											type="button"
											disabled={axesLocked}
											onClick={() =>
												selectProductType(option.type)
											}
											className={`p-4 rounded-xl border text-start transition-all ${
												active
													? "border-accent-brand bg-accent-brand/5 ring-1 ring-accent-brand shadow-2xs"
													: "border-border bg-secondary/20 hover:bg-secondary/40"
											} ${
												axesLocked
													? "opacity-60 cursor-not-allowed"
													: ""
											}`}
										>
											<div className="flex items-center justify-between mb-1.5">
												<span className="text-xs font-bold text-foreground">
													{option.label}
												</span>
												<span
													className={`size-4 rounded-full border flex items-center justify-center ${
														active
															? "border-accent-brand bg-accent-brand text-white"
															: "border-border"
													}`}
												>
													{active && (
														<Check className="size-2.5 stroke-[3]" />
													)}
												</span>
											</div>
											<p className="text-[11px] text-muted-foreground leading-relaxed">
												{option.description}
											</p>
										</button>
									);
								})}
							</div>
							{axesLocked && (
								<p className="text-[11px] text-amber-600 dark:text-amber-400 font-medium">
									{t("products.form.axesLocked")}
								</p>
							)}
						</FormItem>
					)}
				/>
			</div>
		</Form>
	);
}
