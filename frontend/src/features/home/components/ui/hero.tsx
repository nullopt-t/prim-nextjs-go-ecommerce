"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/navigation";
import {
  ChevronLeft,
  ChevronRight,
  Sparkles,
  ArrowRight,
  Headphones,
  Laptop,
  Watch,
  Pause,
  Play,
} from "lucide-react";

interface SlideData {
  id: string;
  tagKey: string;
  titleKey: string;
  highlightKey: string;
  descKey: string;
  ctaKey: string;
  link: string;
  badgeIcon: any;
  visualIcon: any;
  accentColor: string;
  accentBg: string;
  accentBorder: string;
  badgeColor: string;
  decorGradient: string;
}

const SLIDES: SlideData[] = [
  {
    id: "slide1",
    tagKey: "slides.slide1.tag",
    titleKey: "slides.slide1.title",
    highlightKey: "slides.slide1.highlight",
    descKey: "slides.slide1.desc",
    ctaKey: "slides.slide1.cta",
    link: "/products",
    badgeIcon: Sparkles,
    visualIcon: Headphones,
    accentColor: "text-accent-brand",
    accentBg: "bg-accent-brand",
    accentBorder: "border-accent-brand/40",
    badgeColor: "bg-accent-brand/15 text-accent-brand border-accent-brand/30",
    decorGradient: "from-accent-brand/20 via-transparent to-transparent",
  },
  {
    id: "slide2",
    tagKey: "slides.slide2.tag",
    titleKey: "slides.slide2.title",
    highlightKey: "slides.slide2.highlight",
    descKey: "slides.slide2.desc",
    ctaKey: "slides.slide2.cta",
    link: "/products",
    badgeIcon: Sparkles,
    visualIcon: Laptop,
    accentColor: "text-sky-500 dark:text-sky-400",
    accentBg: "bg-sky-500",
    accentBorder: "border-sky-500/40",
    badgeColor: "bg-sky-500/15 text-sky-600 dark:text-sky-400 border-sky-500/30",
    decorGradient: "from-sky-500/20 via-transparent to-transparent",
  },
  {
    id: "slide3",
    tagKey: "slides.slide3.tag",
    titleKey: "slides.slide3.title",
    highlightKey: "slides.slide3.highlight",
    descKey: "slides.slide3.desc",
    ctaKey: "slides.slide3.cta",
    link: "/products",
    badgeIcon: Sparkles,
    visualIcon: Watch,
    accentColor: "text-emerald-500 dark:text-emerald-400",
    accentBg: "bg-emerald-500",
    accentBorder: "border-emerald-500/40",
    badgeColor: "bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 border-emerald-500/30",
    decorGradient: "from-emerald-500/20 via-transparent to-transparent",
  },
];

const AUTOPLAY_INTERVAL = 6000;

export default function Hero() {
  const t = useTranslations("home.hero");
  const [currentIndex, setCurrentIndex] = useState(0);
  const [isPlaying, setIsPlaying] = useState(true);
  const timerRef = useRef<NodeJS.Timeout | null>(null);

  const nextSlide = useCallback(() => {
    setCurrentIndex((prev) => (prev + 1) % SLIDES.length);
  }, []);

  const prevSlide = useCallback(() => {
    setCurrentIndex((prev) => (prev - 1 + SLIDES.length) % SLIDES.length);
  }, []);

  useEffect(() => {
    if (!isPlaying) return;
    timerRef.current = setInterval(() => {
      nextSlide();
    }, AUTOPLAY_INTERVAL);

    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [isPlaying, nextSlide, currentIndex]);

  const activeSlide = SLIDES[currentIndex];
  const VisualIcon = activeSlide.visualIcon;
  const BadgeIcon = activeSlide.badgeIcon;

  return (
    <div
      className="relative w-full rounded-3xl overflow-hidden border border-border/70 bg-card shadow-sm hover:shadow-md transition-shadow group/hero"
      onMouseEnter={() => setIsPlaying(false)}
      onMouseLeave={() => setIsPlaying(true)}
      aria-roledescription="carousel"
      aria-label="Promotions Carousel"
    >
      {/* Dynamic ambient backdrop light */}
      <div
        className={`pointer-events-none absolute -top-32 -right-32 size-96 rounded-full blur-3xl opacity-30 transition-all duration-700 bg-linear-to-br ${activeSlide.decorGradient}`}
      />
      <div
        className={`pointer-events-none absolute -bottom-32 -left-32 size-96 rounded-full blur-3xl opacity-20 transition-all duration-700 bg-linear-to-tr ${activeSlide.decorGradient}`}
      />

      {/* Main slide display */}
      <div className="relative z-10 px-5 py-6 sm:px-8 sm:py-8 md:px-10 md:py-10 flex flex-col md:flex-row items-center justify-between gap-6 md:gap-8 min-h-[300px] md:min-h-[340px]">
        {/* Text column */}
        <div className="flex-1 max-w-xl text-left rtl:text-right">
          {/* Tag pill */}
          <div
            className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full border text-xs font-semibold uppercase tracking-wider mb-2.5 transition-colors ${activeSlide.badgeColor}`}
          >
            <BadgeIcon className="size-3.5" />
            <span>{t(activeSlide.tagKey)}</span>
          </div>

          {/* Heading */}
          <h1 className="text-xl xs:text-2xl sm:text-3xl md:text-4xl font-black text-foreground tracking-tight leading-tight mb-2">
            <span>{t(activeSlide.titleKey)} </span>
            <span className={`block sm:inline ${activeSlide.accentColor}`}>
              {t(activeSlide.highlightKey)}
            </span>
          </h1>

          {/* Description */}
          <p className="text-muted-foreground text-xs sm:text-sm leading-relaxed mb-5 max-w-lg line-clamp-2 sm:line-clamp-3">
            {t(activeSlide.descKey)}
          </p>

          {/* CTA Actions */}
          <div className="flex flex-wrap items-center gap-3">
            <Link href={activeSlide.link} className="inline-block">
              <div
                className={`px-5 py-2.5 rounded-xl font-bold text-xs sm:text-sm text-white shadow-xs hover:shadow-md transition-all flex items-center gap-2 cursor-pointer hover:-translate-y-0.5 ${activeSlide.accentBg}`}
              >
                <span>{t(activeSlide.ctaKey)}</span>
                <ArrowRight className="size-4 rtl:rotate-180" />
              </div>
            </Link>

            <Link href="/products" className="inline-block">
              <div className="px-4 py-2.5 rounded-xl font-semibold text-xs sm:text-sm text-foreground bg-secondary/80 hover:bg-secondary border border-border/60 transition-all hover:-translate-y-0.5 cursor-pointer">
                {t("viewDeals")}
              </div>
            </Link>
          </div>
        </div>

        {/* Visual column with interactive highlight card */}
        <div className="w-full md:w-auto shrink-0 flex items-center justify-center">
          <div className="relative size-44 sm:size-52 md:size-56 rounded-2xl bg-secondary/40 border border-border/60 p-4 flex flex-col items-center justify-center text-center shadow-inner group/card">
            {/* Subtle inner decorative glow */}
            <div className="absolute inset-2.5 rounded-xl border border-dashed border-border/50 pointer-events-none" />

            <div className="relative z-10 size-16 sm:size-20 md:size-24 rounded-2xl bg-card border border-border/80 shadow-xs flex items-center justify-center text-foreground transition-transform duration-500 group-hover/card:scale-105">
              <VisualIcon className={`size-8 sm:size-10 md:size-12 transition-colors ${activeSlide.accentColor}`} />
            </div>

            <div className="relative z-10 mt-3 text-center">
              <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest block">
                PRIM Exclusive
              </span>
              <p className="text-xs font-semibold text-foreground mt-0.5">
                Curated Collection
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Bottom controls bar: Slide indicators, prev/next arrows & play/pause */}
      <div className="relative z-10 px-5 py-2.5 bg-secondary/30 border-t border-border/50 flex items-center justify-between gap-4">
        {/* Slide indicators / pagination */}
        <div className="flex items-center gap-2">
          {SLIDES.map((slide, idx) => {
            const isActive = idx === currentIndex;
            return (
              <button
                key={slide.id}
                onClick={() => setCurrentIndex(idx)}
                aria-label={`Go to slide ${idx + 1}`}
                className={`h-2 rounded-full transition-all duration-300 ${
                  isActive
                    ? `w-8 ${slide.accentBg}`
                    : "w-2 bg-muted-foreground/30 hover:bg-muted-foreground/50"
                }`}
              />
            );
          })}
        </div>

        {/* Carousel controls */}
        <div className="flex items-center gap-1.5 sm:gap-2">
          <button
            onClick={() => setIsPlaying((prev) => !prev)}
            aria-label={isPlaying ? "Pause autoplay" : "Play autoplay"}
            className="size-8 rounded-lg border border-border/60 hover:bg-secondary flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
          >
            {isPlaying ? <Pause className="size-3.5" /> : <Play className="size-3.5 ml-0.5" />}
          </button>

          <button
            onClick={prevSlide}
            aria-label="Previous slide"
            className="size-8 rounded-lg border border-border/60 hover:bg-secondary flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
          >
            <ChevronLeft className="size-4 rtl:rotate-180" />
          </button>

          <button
            onClick={nextSlide}
            aria-label="Next slide"
            className="size-8 rounded-lg border border-border/60 hover:bg-secondary flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
          >
            <ChevronRight className="size-4 rtl:rotate-180" />
          </button>
        </div>
      </div>
    </div>
  );
}
