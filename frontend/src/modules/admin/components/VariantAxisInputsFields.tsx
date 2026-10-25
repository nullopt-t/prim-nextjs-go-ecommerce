"use client";

import { useTranslations } from "next-intl";

interface VariantAxisInputsFieldsProps {
	axes: string[];
	attributes: Record<string, string>;
	definitions: Record<string, any>;
	draftValues: Record<string, string[]>;
	onAttributeChange: (key: string, value: string) => void;
}

export function VariantAxisInputsFields({
	axes,
	attributes,
	definitions,
	draftValues,
	onAttributeChange,
}: VariantAxisInputsFieldsProps) {
	const t = useTranslations("admin");

	return (
		<div className="space-y-3 pt-3 border-t border-border">
			<div className="flex items-center justify-between">
				<label className="text-xs font-bold text-foreground">
					{t("products.variants.variationAxes")} *
				</label>
				<span className="text-[11px] text-muted-foreground">
					{t("products.wizard.step4VariableHint")}
				</span>
			</div>

			<div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3">
				{axes.map((key) => {
					const attribute = definitions[key];
					const label = attribute?.label || key;
					const draftVals = draftValues[key] || [];
					const currentValue = attributes[key] ?? "";

					return (
						<div key={key} className="space-y-1.5">
							<label className="text-[11px] font-semibold text-foreground">
								{label} *
							</label>
							<input
								type="text"
								list={`datalist-${key}`}
								value={currentValue}
								onChange={(e) => onAttributeChange(key, e.target.value)}
								placeholder={`Type ${label} value...`}
								className="w-full rounded-xl border border-border bg-background px-3 py-2 text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
							/>

							{draftVals.length > 0 && (
								<>
									<datalist id={`datalist-${key}`}>
										{draftVals.map((opt, optIdx) => (
											<option key={`dl-${key}-${opt}-${optIdx}`} value={opt} />
										))}
									</datalist>

									<div className="flex flex-wrap gap-1">
										{draftVals.map((opt, optIdx) => (
											<button
												key={`chip-${key}-${opt}-${optIdx}`}
												type="button"
												onClick={() => onAttributeChange(key, opt)}
												className={`px-2 py-0.5 rounded-md text-[10px] font-medium border transition-all ${
													currentValue === opt
														? "border-accent-brand bg-accent-brand/20 text-accent-brand font-bold"
														: "border-border bg-secondary/30 text-muted-foreground hover:bg-secondary hover:text-foreground"
												}`}
											>
												{opt}
											</button>
										))}
									</div>
								</>
							)}
						</div>
					);
				})}
			</div>
		</div>
	);
}
