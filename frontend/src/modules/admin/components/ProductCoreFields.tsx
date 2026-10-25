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
	FormTextarea,
} from "@/shared/ui/form/form";

export interface ProductCoreValues {
	name: string;
	categoryId: string;
	description: string;
}

interface ProductCoreFieldsProps {
	value: ProductCoreValues;
	categories: Array<{ id: string; name: string }>;
	currentCategoryName: string;
	onChange: (
		patch: Partial<ProductCoreValues> & { category?: string },
	) => void;
}

export function ProductCoreFields({
	value,
	categories,
	currentCategoryName,
	onChange,
}: ProductCoreFieldsProps) {
	const t = useTranslations("admin");
	const schema = useMemo(
		() =>
			z.object({
				name: z
					.string()
					.trim()
					.min(1, t("products.validation.titleRequired")),
				categoryId: z
					.string()
					.min(1, t("products.validation.categoryRequired")),
				description: z.string(),
			}),
		[t],
	);

	const form = useForm<ProductCoreValues>({
		resolver: zodResolver(schema),
		defaultValues: value,
		mode: "onTouched",
	});

	const inputClass =
		"bg-secondary/30 font-sans focus:border-accent-brand focus:ring-accent-brand/30";

	return (
		<Form {...form}>
			<div className="space-y-4">
				<FormField
					control={form.control}
					name="name"
					render={({ field }) => (
						<FormItem>
							<FormLabel className="font-bold">
								{t("products.form.title")} *
							</FormLabel>
							<FormControl>
								<FormInput
									{...field}
									className={inputClass}
									placeholder={t("products.form.titlePlaceholder")}
									onChange={(e) => {
										field.onChange(e);
										onChange({ name: e.target.value });
									}}
								/>
							</FormControl>
							<FormMessage />
						</FormItem>
					)}
				/>

				<FormField
					control={form.control}
					name="categoryId"
					render={({ field }) => (
						<FormItem>
							<FormLabel className="font-bold">
								{t("products.form.category")} *
							</FormLabel>
							<FormControl>
								<select
									{...field}
									className="w-full px-3.5 py-2.5 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
									onChange={(e) => {
										field.onChange(e);
										const cat = categories.find(
											(c) => c.id === e.target.value,
										);
										onChange({
											categoryId: e.target.value,
											category: cat?.name || currentCategoryName,
										});
									}}
								>
									{categories.map((category) => (
										<option
											key={`cat-opt-${category.id}`}
											value={category.id}
										>
											{category.name}
										</option>
									))}
								</select>
							</FormControl>
							<p className="text-[11px] text-muted-foreground">
								{t("products.wizard.selectedCategoryNotice", {
									category: currentCategoryName,
								})}
							</p>
							<FormMessage />
						</FormItem>
					)}
				/>

				<FormField
					control={form.control}
					name="description"
					render={({ field }) => (
						<FormItem>
							<FormLabel className="font-bold">
								{t("products.form.description")}
							</FormLabel>
							<FormControl>
								<FormTextarea
									{...field}
									rows={4}
									className={inputClass}
									placeholder={t(
										"products.form.descriptionPlaceholder",
									)}
									onChange={(e) => {
										field.onChange(e);
										onChange({ description: e.target.value });
									}}
								/>
							</FormControl>
							<FormMessage />
						</FormItem>
					)}
				/>
			</div>
		</Form>
	);
}
