"use client";

import { useRef, type ReactNode } from "react";
import { Link } from "@/shared/i18n/navigation";
import { ArrowRight, ChevronLeft, ChevronRight } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";
import { toCardDetails } from "./productCardMapper";
import type { ProductDetailsEntity } from "@/modules/catalog/types";
import { ProductsCard } from "@/shared/ui/productsCard";

/** Horizontal scrolling product shelf shared by the merchandising rows. */
export default function ProductShelf({
	title,
	subtitle,
	badge,
	link,
	products,
}: {
	title: string;
	subtitle?: string;
	badge?: ReactNode;
	link: string;
	products: ProductDetailsEntity[];
}) {
	const scrollRef = useRef<HTMLDivElement>(null);
	const locale = useLocale();
	const t = useTranslations("home");

	const scroll = (direction: "left" | "right") => {
		if (!scrollRef.current) return;
		// In RTL the horizontal scroll axis inverts, so flip the delta to keep
		// "next" moving toward the end of the list.
		const scrollAmount = (direction === "left" ? -400 : 400) * (locale === "ar" ? -1 : 1);
		scrollRef.current.scrollBy({ left: scrollAmount, behavior: "smooth" });
	};

	return (
		<section className="bg-card border border-border hover:border-accent-brand/40 rounded-2xl p-3.5 sm:p-5 relative group shadow-2xs transition-colors">
			<div className="flex items-center justify-between mb-3 sm:mb-4">
				<div className="text-start flex items-center gap-2">
					{badge}
					<div>
						<h2 dir="auto" className="text-base sm:text-lg md:text-xl font-bold text-foreground">
							{title}
						</h2>
						{subtitle && (
							<p className="text-[11px] sm:text-xs text-muted-foreground mt-0.5">{subtitle}</p>
						)}
					</div>
				</div>
				<Link
					href={link}
					className="text-xs font-semibold text-accent-brand hover:underline hover:underline-offset-4 inline-flex items-center gap-1.5 transition-colors shrink-0"
				>
					<span>{t("categories.seeAll")}</span>
					<ArrowRight className="size-3.5 rtl:rotate-180" />
				</Link>
			</div>

			<button
				type="button"
				onClick={() => scroll("left")}
				aria-label="scroll-left"
				className="flex absolute ltr:left-2 rtl:right-2 top-1/2 -translate-y-1/2 z-10 size-11 rounded-full bg-background/90 backdrop-blur-sm border border-border shadow-md items-center justify-center text-foreground hover:bg-accent hover:text-accent-brand opacity-100 sm:opacity-0 sm:group-hover:opacity-100 sm:group-focus-within:opacity-100 transition-opacity cursor-pointer"
			>
				<ChevronLeft className="size-4 rtl:rotate-180" />
			</button>
			<button
				type="button"
				onClick={() => scroll("right")}
				aria-label="scroll-right"
				className="flex absolute ltr:right-2 rtl:left-2 top-1/2 -translate-y-1/2 z-10 size-11 rounded-full bg-background/90 backdrop-blur-sm border border-border shadow-md items-center justify-center text-foreground hover:bg-accent hover:text-accent-brand opacity-100 sm:opacity-0 sm:group-hover:opacity-100 sm:group-focus-within:opacity-100 transition-opacity cursor-pointer"
			>
				<ChevronRight className="size-4 rtl:rotate-180" />
			</button>

			<div
				ref={scrollRef}
				className="flex gap-3 sm:gap-4 overflow-x-auto no-scrollbar scroll-smooth pb-1"
				style={{ scrollbarWidth: "none", msOverflowStyle: "none" }}
			>
				{products.map((product) => (
					<div key={String(product.slug || product.id)} className="w-[180px] sm:w-[200px] md:w-[220px] shrink-0">
						<ProductsCard cardDetails={toCardDetails(product)} />
					</div>
				))}
			</div>
		</section>
	);
}
