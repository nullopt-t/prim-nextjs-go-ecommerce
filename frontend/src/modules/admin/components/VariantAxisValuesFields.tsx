"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { Loader2, Sparkles, X } from "lucide-react";
import { notify } from "@/shared/lib/notify";

interface VariantAxisValuesFieldsProps {
	axes: string[];
	values: Record<string, string[]>;
	definitions: Record<string, any>;
	stock: string;
	onAddValue: (key: string, value: string) => void;
	onRemoveValue: (key: string, value: string) => void;
	onChangeStock: (value: string) => void;
	onCreateMissing: () => void;
	creating: boolean;
	ready: boolean;
	count: number;
}

export function VariantAxisValuesFields({
	axes,
	values,
	definitions,
	stock,
	onAddValue,
	onRemoveValue,
	onChangeStock,
	onCreateMissing,
	creating,
	ready,
	count,
}: VariantAxisValuesFieldsProps) {
	const t = useTranslations("admin");
	const [inputs, setInputs] = useState<Record<string, string>>({});

	const handleAddAxisValue = (key: string) => {
		const value = (inputs[key] || "").trim();
		if (!value) return;
		const existingValues = values[key] || [];
		if (existingValues.includes(value)) {
			notify.error(t("products.variants.axisValueExists", { value }));
			return;
		}

		const definition = definitions[key];
		const allowedValues: string[] = Array.isArray(definition?.allowedValues)
			? definition.allowedValues.map((allowed: any) => String(allowed))
			: [];
		if (
			allowedValues.length > 0 &&
			!allowedValues.some((allowed) => allowed.toLowerCase() === value.toLowerCase())
		) {
			notify.error(
				t("products.variants.axisValueNotAllowed", {
					value,
					allowed: allowedValues.join(", "),
				}),
			);
			return;
		}

		onAddValue(key, value);
		setInputs((prev) => ({ ...prev, [key]: "" }));
	};

	return (
		<div className="rounded-2xl border border-border bg-secondary/15 p-5 space-y-4">
			<div className="flex items-center justify-between">
				<div>
					<h3 className="text-xs font-bold text-foreground uppercase tracking-wider">
						{t("products.variants.variationAxes")}
					</h3>
					<p className="text-[11px] text-muted-foreground mt-0.5">
						{t("products.wizard.step3OptionsHint")}
					</p>
				</div>
				<span className="text-xs font-semibold px-2.5 py-1 rounded-lg bg-accent-brand/10 text-accent-brand">
					{t("products.variants.optionsCount", { count: axes.length })}
				</span>
			</div>

			<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
				{axes.map((key) => {
					const definition = definitions[key];
					const axisValues = values[key] || [];

					return (
						<div key={key} className="rounded-xl border border-border bg-background p-4 space-y-3 shadow-2xs">
							<div className="flex items-center justify-between">
								<span className="text-xs font-bold text-foreground">
									{definition?.label || key}
								</span>
								<span className="text-[10px] text-muted-foreground">
									{t("products.variants.valueCount", { count: axisValues.length })}
								</span>
							</div>

							{/* Value chips */}
							<div className="flex flex-wrap gap-1.5 min-h-[30px] items-center">
								{axisValues.length === 0 ? (
									<span className="text-[11px] text-muted-foreground italic">
										{t("products.variants.noAxisValues")}
									</span>
								) : (
									axisValues.map((value) => (
										<span
											key={value}
											className="inline-flex items-center gap-1 rounded-md border border-border bg-secondary/40 px-2.5 py-1 text-xs font-medium text-foreground"
										>
											{value}
											<button
												type="button"
												onClick={() => onRemoveValue(key, value)}
												className="text-muted-foreground hover:text-destructive"
											>
												<X className="size-3" />
											</button>
										</span>
									))
								)}
							</div>

							{/* Add value input */}
							<div className="flex gap-2">
								<input
									type="text"
									value={inputs[key] || ""}
									onChange={(e) =>
										setInputs((prev) => ({ ...prev, [key]: e.target.value }))
									}
									onKeyDown={(e) => {
										if (e.key === "Enter") {
											e.preventDefault();
											handleAddAxisValue(key);
										}
									}}
									placeholder={t("products.variants.addValuePlaceholder")}
									className="min-w-0 flex-1 rounded-xl border border-border bg-background px-3 py-1.5 text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
								/>
								<button
									type="button"
									onClick={() => handleAddAxisValue(key)}
									className="rounded-xl border border-border px-3 py-1.5 text-xs font-semibold text-foreground hover:bg-secondary shrink-0"
								>
									{t("products.variants.addValueAction")}
								</button>
							</div>
						</div>
					);
				})}
			</div>

			<div className="flex flex-wrap items-end gap-3 pt-2">
				<div className="space-y-1">
					<label className="text-[11px] font-semibold text-foreground">
						{t("products.variants.matrixStock")}
					</label>
					<input
						type="number"
						min="0"
						value={stock}
						onChange={(e) => onChangeStock(e.target.value)}
						className="w-32 rounded-xl border border-border bg-background px-3 py-2 text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
					/>
				</div>

				<div className="space-y-1">
					<div className="text-[11px] font-semibold text-foreground">
						{t("products.variants.combinations", { count })}
					</div>
					<div className="text-[11px] text-muted-foreground">
						{ready ? t("products.variants.combinationsReady") : t("products.variants.combinationsMissing")}
					</div>
				</div>

				<button
					type="button"
					onClick={onCreateMissing}
					disabled={creating || !ready}
					className="flex items-center justify-center gap-2 rounded-xl bg-accent-brand px-5 py-2.5 text-xs font-semibold text-white hover:bg-accent-brand/90 disabled:opacity-60 shadow-xs"
				>
					{creating && <Loader2 className="size-3.5 animate-spin" />}
					<Sparkles className="size-3.5" />
					<span>{t("products.variants.createMissing")}</span>
				</button>
			</div>
		</div>
	);
}
