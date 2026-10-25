"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { Loader2, Plus, X } from "lucide-react";
import { notify } from "@/shared/lib/notify";

export interface ProductOptionAxesValues {
	variationAxes: string[];
}

interface ProductOptionAxesFieldsProps {
	value: ProductOptionAxesValues;
	categoryAttributes: Array<{ key: string; label?: string }>;
	locked: boolean;
	creating: boolean;
	onChange: (patch: Partial<ProductOptionAxesValues>) => void;
	onCreateAxis: (label: string) => void;
}

export function ProductOptionAxesFields({
	value,
	categoryAttributes,
	locked,
	creating,
	onChange,
	onCreateAxis,
}: ProductOptionAxesFieldsProps) {
	const t = useTranslations("admin");
	const [formAxisToAdd, setFormAxisToAdd] = useState("");
	const [newAxisName, setNewAxisName] = useState("");

	const inactiveCandidates = categoryAttributes.filter(
		(attribute) => !value.variationAxes.includes(attribute.key),
	);

	const handleAddAxis = (key: string) => {
		if (locked || value.variationAxes.includes(key)) return;
		onChange({ variationAxes: [...value.variationAxes, key] });
		setFormAxisToAdd("");
	};

	const handleRemoveAxis = (key: string) => {
		if (locked) return;
		onChange({
			variationAxes: value.variationAxes.filter((axis) => axis !== key),
		});
	};

	const handleCreateAxis = () => {
		if (!newAxisName.trim() || locked || creating) return;
		onCreateAxis(newAxisName.trim());
		setNewAxisName("");
	};

	return (
		<div className="space-y-4 rounded-xl border border-border bg-secondary/20 p-5 max-w-2xl">
			<div>
				<label className="text-xs font-bold text-foreground">
					{t("products.form.variationAxes")} *
				</label>
				<p className="text-[11px] text-muted-foreground mt-0.5">
					{t("products.form.variationAxesHint")}
				</p>
			</div>

			{/* Selected Options Pills */}
			<div className="flex flex-wrap gap-2 min-h-[36px] items-center">
				{value.variationAxes.length === 0 ? (
					<span className="text-xs text-muted-foreground italic">
						{t("products.form.noAxes")}
					</span>
				) : (
					value.variationAxes.map((key) => {
						const definition = categoryAttributes.find(
							(attribute) => attribute.key === key,
						);
						return (
							<span
								key={key}
								className="flex items-center gap-2 rounded-full border border-accent-brand/30 bg-accent-brand/10 px-3 py-1.5 text-xs font-semibold text-accent-brand"
							>
								{definition?.label || key}
								{!locked && (
									<button
										type="button"
										onClick={() => handleRemoveAxis(key)}
										className="rounded-full p-0.5 hover:bg-accent-brand/20 transition-all"
										title={t("products.variants.removeAxis")}
									>
										<X className="size-3" />
									</button>
								)}
							</span>
						);
					})
				)}
			</div>

			{/* Dropdown to add from existing category options */}
			{!locked && inactiveCandidates.length > 0 && (
				<div className="flex items-center gap-2 pt-2 border-t border-border">
					<select
						value={formAxisToAdd}
						onChange={(e) => setFormAxisToAdd(e.target.value)}
						className="min-w-0 flex-1 rounded-xl border border-border bg-background px-3 py-2 text-xs text-foreground"
					>
						<option value="">{t("products.variants.addAxis")}</option>
						{inactiveCandidates.map((attribute) => (
							<option key={attribute.key} value={attribute.key}>
								{attribute.label || attribute.key}
							</option>
						))}
					</select>
					<button
						type="button"
						onClick={() => handleAddAxis(formAxisToAdd)}
						disabled={!formAxisToAdd}
						className="px-4 py-2 rounded-xl border border-border text-xs font-semibold text-foreground hover:bg-secondary disabled:opacity-50 shrink-0"
					>
						{t("products.variants.addAxisAction")}
					</button>
				</div>
			)}

			{/* Instant Inline Option Creator & Presets */}
			{!locked && (
				<div className="space-y-3 pt-3 border-t border-border">
					<div className="space-y-1">
						<span className="text-xs font-bold text-foreground block">
							{t("products.wizard.quickAddAxis")}
						</span>
						{categoryAttributes.length === 0 && (
							<p className="text-[11px] text-amber-600 dark:text-amber-400 font-medium">
								{t("products.wizard.noAttributesNotice")}
							</p>
						)}
					</div>

					{/* Presets */}
					<div className="flex flex-wrap gap-2">
						{[
							{ label: t("products.wizard.presetColor"), key: "color" },
							{ label: t("products.wizard.presetSize"), key: "size" },
							{ label: t("products.wizard.presetStorage"), key: "storage" },
							{ label: t("products.wizard.presetMaterial"), key: "material" },
						].map((preset) => {
							const isAdded = value.variationAxes.includes(preset.key);
							return (
								<button
									key={preset.key}
									type="button"
									disabled={isAdded || creating}
									onClick={() => onCreateAxis(preset.label)}
									className={`px-3 py-1.5 rounded-lg border text-xs font-medium transition-all ${
										isAdded
											? "border-transparent bg-secondary/40 text-muted-foreground opacity-50 cursor-default"
											: "border-dashed border-border bg-background text-foreground hover:border-accent-brand/50 hover:text-accent-brand hover:bg-accent-brand/5"
									}`}
								>
									{isAdded ? `✓ ${preset.label}` : `+ ${preset.label}`}
								</button>
							);
						})}
					</div>

					{/* Custom Option Input */}
					<div className="flex items-center gap-2">
						<input
							type="text"
							value={newAxisName}
							onChange={(e) => setNewAxisName(e.target.value)}
							onKeyDown={(e) => {
								if (e.key === "Enter") {
									e.preventDefault();
									handleCreateAxis();
								}
							}}
							placeholder={t("products.wizard.quickAddAxisPlaceholder")}
							className="min-w-0 flex-1 rounded-xl border border-border bg-background px-3 py-2 text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
						/>
						<button
							type="button"
							onClick={handleCreateAxis}
							disabled={!newAxisName.trim() || creating}
							className="flex items-center gap-1.5 px-4 py-2 rounded-xl bg-accent-brand text-white text-xs font-semibold hover:bg-accent-brand/90 disabled:opacity-50 shrink-0 transition-all shadow-2xs"
						>
							{creating && <Loader2 className="size-3 animate-spin" />}
							<Plus className="size-3" />
							{t("products.wizard.createAxisAction")}
						</button>
					</div>
				</div>
			)}
		</div>
	);
}
