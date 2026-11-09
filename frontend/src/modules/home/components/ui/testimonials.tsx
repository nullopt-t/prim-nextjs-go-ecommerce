"use client";

import { BadgeCheck, Quote } from "lucide-react";
import { useTranslations } from "next-intl";
import { Stars } from "@/shared/ui/stars";

export default function Testimonials() {
	const t = useTranslations("home.testimonials");
	const quotes = [
		{ name: t("item1Name"), text: t("item1Quote"), stars: 5 },
		{ name: t("item2Name"), text: t("item2Quote"), stars: 5 },
		{ name: t("item3Name"), text: t("item3Quote"), stars: 4 },
	];

	return (
		<section className="space-y-3">
			<div className="text-start">
				<h2 className="text-base sm:text-lg md:text-xl font-bold text-foreground">{t("title")}</h2>
				<p className="text-[11px] sm:text-xs text-muted-foreground mt-0.5">{t("subtitle")}</p>
			</div>
			<div className="grid grid-cols-1 md:grid-cols-3 gap-3 sm:gap-4">
				{quotes.map((quote) => (
					<figure
						key={quote.name}
						className="flex flex-col gap-3 rounded-2xl border border-border/80 bg-card p-4 sm:p-5 shadow-2xs"
					>
						<div className="flex items-center justify-between gap-2">
							<Stars starsNum={quote.stars} />
							<Quote className="size-5 text-accent-brand/40" aria-hidden />
						</div>
						<blockquote className="text-sm leading-relaxed text-foreground">{quote.text}</blockquote>
						<figcaption className="mt-auto flex items-center gap-1.5 text-xs font-semibold text-foreground">
							<span dir="auto">{quote.name}</span>
							<span className="inline-flex items-center gap-1 font-medium text-emerald-600 dark:text-emerald-400">
								<BadgeCheck className="size-3.5" aria-hidden />
								{t("verifiedBuyer")}
							</span>
						</figcaption>
					</figure>
				))}
			</div>
		</section>
	);
}
