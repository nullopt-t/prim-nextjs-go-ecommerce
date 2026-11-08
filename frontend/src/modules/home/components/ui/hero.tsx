"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useTranslations, useLocale } from "next-intl";
import { Link } from "@/shared/i18n/navigation";
import { ChevronLeft, ChevronRight, ArrowRight, Pause, Play } from "lucide-react";
import type { ProductDetailsEntity } from "@/modules/catalog/types";
import { useCurrency } from "@/modules/settings/model/publicSettingsContext";
import { formatMoney } from "@/shared/lib/money";
import { normalizeMediaUrl } from "@/shared/lib/media";

const AUTOPLAY_INTERVAL = 6000;

function productName(product: ProductDetailsEntity, locale: string): string {
	const name =
		typeof product.product === "object"
			? locale === "ar"
				? product.product?.ar || product.product?.en
				: product.product?.en || product.product?.ar
			: product.product;
	return (typeof name === "string" && name.trim()) || product.title || "";
}

export default function Hero({ products = [] }: { products?: ProductDetailsEntity[] }) {
	const t = useTranslations("home.hero");
	const locale = useLocale();
	const currency = useCurrency();
	const [currentIndex, setCurrentIndex] = useState(0);
	const [isPlaying, setIsPlaying] = useState(true);
	const timerRef = useRef<NodeJS.Timeout | null>(null);

	const slides = (products ?? [])
		.filter((product) => product.img)
		.slice(0, 3);

	const nextSlide = useCallback(() => {
		setCurrentIndex((prev) => (prev + 1) % slides.length);
	}, [slides.length]);

	const prevSlide = useCallback(() => {
		setCurrentIndex((prev) => (prev - 1 + slides.length) % slides.length);
	}, [slides.length]);

	useEffect(() => {
		if (!isPlaying || slides.length < 2) return;
		if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;
		timerRef.current = setInterval(() => {
			nextSlide();
		}, AUTOPLAY_INTERVAL);

		return () => {
			if (timerRef.current) clearInterval(timerRef.current);
		};
	}, [isPlaying, nextSlide, slides.length, currentIndex]);

	if (slides.length === 0) return null;

	const activeSlide = slides[currentIndex % slides.length];
	const activeName = productName(activeSlide, locale);
	const activeHref = `/products/${activeSlide.slug || activeSlide.id}`;

	return (
		<div
			className="relative w-full overflow-hidden rounded-2xl border border-border/60 group/hero"
			onMouseEnter={() => setIsPlaying(false)}
			onMouseLeave={() => setIsPlaying(true)}
			onFocus={() => setIsPlaying(false)}
			onBlur={() => setIsPlaying(true)}
			aria-roledescription="carousel"
			aria-label={t("billboard")}
		>
			<div className="relative flex min-h-[290px] flex-col items-center justify-between gap-6 px-6 py-8 sm:min-h-[320px] sm:px-10 sm:py-10 md:flex-row md:gap-12 md:px-14 md:py-12">
				<div className="z-10 max-w-lg flex-1 text-start">
					<div className="mb-3 inline-flex items-center gap-2">
						<span className="rounded bg-secondary px-2.5 py-0.5 text-[11px] font-bold uppercase tracking-wider text-foreground">
							{t("featured")}
						</span>
						{activeSlide.category ? (
							<span className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
								{activeSlide.category.replace(/-/g, " ")}
							</span>
						) : null}
					</div>

					<h1 className="mb-2.5 text-2xl font-bold leading-tight tracking-tight text-foreground sm:text-3xl md:text-4xl">
						{activeName}
					</h1>

					<p className="mb-6 max-w-md text-sm leading-relaxed text-muted-foreground">
						{formatMoney(activeSlide.price, currency)}
						{activeSlide.oldPrice && activeSlide.oldPrice > activeSlide.price
							? ` · ${formatMoney(activeSlide.oldPrice, currency)}`
							: ""}
					</p>

					<div className="flex items-center gap-3">
						<Link
							href={activeHref}
							className="inline-flex items-center gap-2 rounded-full bg-foreground px-6 py-2.5 text-sm font-semibold text-background shadow-xs transition-transform hover:-translate-y-0.5"
						>
							<span>{t("viewProduct")}</span>
							<ArrowRight className="size-4 rtl:rotate-180" />
						</Link>
					</div>
				</div>

				<div className="z-10 flex w-full shrink-0 items-center justify-center md:w-auto">
					<div className="relative flex size-56 items-center justify-center overflow-hidden rounded-2xl border border-border/50 bg-background p-3 shadow-md transition-all duration-300 group-hover/hero:shadow-lg sm:size-64 md:size-72">
						<img
							key={activeSlide.id}
							src={normalizeMediaUrl(activeSlide.img)}
							alt={activeName}
							className="h-full w-full rounded-xl object-cover transition-transform duration-500 ease-out group-hover/hero:scale-105"
						/>
					</div>
				</div>
			</div>

			{slides.length > 1 ? (
				<div className="relative flex items-center justify-between gap-4 border-t border-border/40 bg-background/60 px-6 py-2.5 backdrop-blur-xs dark:bg-background/40">
					<div className="flex items-center gap-2">
						{slides.map((slide, idx) => {
							const isActive = idx === currentIndex % slides.length;
							return (
								<button
									key={slide.id}
									type="button"
									onClick={() => setCurrentIndex(idx)}
									aria-label={t("goToSlide", { slide: idx + 1 })}
									aria-current={isActive ? "true" : undefined}
									className="flex min-h-6 min-w-6 items-center justify-center p-1"
								>
									<span
										className={`rounded-full transition-all duration-200 ${
											isActive ? "h-2.5 w-7 bg-foreground" : "h-2.5 w-2 bg-muted-foreground/30 hover:bg-muted-foreground/60"
										}`}
									/>
								</button>
							);
						})}
					</div>

					<div className="flex items-center gap-1.5">
						<button
							type="button"
							onClick={() => setIsPlaying((prev) => !prev)}
							aria-label={isPlaying ? t("pauseCarousel") : t("playCarousel")}
							className="flex h-11 w-11 items-center justify-center rounded-full border border-border bg-background text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
						>
							{isPlaying ? <Pause className="size-4" /> : <Play className="size-4" />}
						</button>

						<button
							type="button"
							onClick={prevSlide}
							aria-label={t("previousSlide")}
							className="flex h-11 w-11 items-center justify-center rounded-full border border-border bg-background text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
						>
							<ChevronLeft className="size-4.5 rtl:rotate-180" />
						</button>

						<button
							type="button"
							onClick={nextSlide}
							aria-label={t("nextSlide")}
							className="flex h-11 w-11 items-center justify-center rounded-full border border-border bg-background text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
						>
							<ChevronRight className="size-4.5 rtl:rotate-180" />
						</button>
					</div>
				</div>
			) : null}
		</div>
	);
}
