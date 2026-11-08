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
import { CustomSelect } from "@/shared/ui/customSelect";

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
								<CustomSelect
									name={field.name}
									ref={field.ref}
									value={field.value ?? ""}
									onChange={(value) => {
										field.onChange(value);
										const cat = categories.find((c) => c.id === value);
										onChange({
										categoryId: value,
										category: cat?.name || currentCategoryName,
										});
									}}
									options={categories.map((category) => ({
										value: category.id,
										label: category.name,
									}))}
									size="sm"
									className="w-full"
/>
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
