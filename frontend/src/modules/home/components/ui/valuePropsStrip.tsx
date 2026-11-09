"use client";

import { Truck, ShieldCheck, RotateCcw, Zap } from "lucide-react";
import { useTranslations } from "next-intl";

export default function ValuePropsStrip() {
	const t = useTranslations("home.valueProps");
	const props = [
		{ icon: Truck, title: t("shippingTitle"), description: t("shippingDescription") },
		{ icon: ShieldCheck, title: t("secureTitle"), description: t("secureDescription") },
		{ icon: RotateCcw, title: t("returnsTitle"), description: t("returnsDescription") },
		{ icon: Zap, title: t("fastTitle"), description: t("fastDescription") },
	];

	return (
		<section aria-label={t("shippingTitle")} className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
			{props.map((prop) => (
				<div
					key={prop.title}
					className="flex items-start gap-3 rounded-2xl border border-border/80 bg-card p-4 shadow-2xs"
				>
					<span className="inline-flex size-10 shrink-0 items-center justify-center rounded-xl bg-accent-brand/10 text-accent-brand">
						<prop.icon className="size-5" />
					</span>
					<div className="text-start">
						<p className="text-sm font-semibold text-foreground">{prop.title}</p>
						<p className="mt-0.5 text-xs leading-relaxed text-muted-foreground">{prop.description}</p>
					</div>
				</div>
			))}
		</section>
	);
}
