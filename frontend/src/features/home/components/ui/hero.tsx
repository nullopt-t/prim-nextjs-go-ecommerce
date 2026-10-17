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
  productImage: string;
  productAlt: string;
  badgeText: string;
  bannerBg: string;
  badgeBg: string;
  ctaClass: string;
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
    productImage: "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?auto=format&fit=crop&q=85&w=800",
    productAlt: "Premium Wireless Headphones",
    badgeText: "DEAL OF THE DAY",
    bannerBg: "bg-[#f7f7f7] dark:bg-secondary/40",
    badgeBg: "bg-red-600 text-white",
    ctaClass: "bg-foreground text-background hover:bg-foreground/90",
  },
  {
    id: "slide2",
    tagKey: "slides.slide2.tag",
    titleKey: "slides.slide2.title",
    highlightKey: "slides.slide2.highlight",
    descKey: "slides.slide2.desc",
    ctaKey: "slides.slide2.cta",
    link: "/products",
    productImage: "https://images.unsplash.com/photo-1587829741301-dc798b83add3?auto=format&fit=crop&q=85&w=800",
    productAlt: "Custom Mechanical Keyboards",
    badgeText: "TOP TECH PICK",
    bannerBg: "bg-[#eef2f6] dark:bg-secondary/40",
    badgeBg: "bg-sky-600 text-white",
    ctaClass: "bg-foreground text-background hover:bg-foreground/90",
  },
  {
    id: "slide3",
    tagKey: "slides.slide3.tag",
    titleKey: "slides.slide3.title",
    highlightKey: "slides.slide3.highlight",
    descKey: "slides.slide3.desc",
    ctaKey: "slides.slide3.cta",
    link: "/products",
    productImage: "https://images.unsplash.com/photo-1523275335684-37898b6baf30?auto=format&fit=crop&q=85&w=800",
    productAlt: "Smart Lifestyle Watch",
    badgeText: "CERTIFIED REFURBISHED",
    bannerBg: "bg-[#f5f5f0] dark:bg-secondary/40",
    badgeBg: "bg-emerald-700 text-white",
    ctaClass: "bg-foreground text-background hover:bg-foreground/90",
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

  return (
    <div
      className={`relative w-full rounded-2xl overflow-hidden border border-border/60 transition-colors duration-500 group/hero ${activeSlide.bannerBg}`}
      onMouseEnter={() => setIsPlaying(false)}
      onMouseLeave={() => setIsPlaying(true)}
      aria-roledescription="carousel"
      aria-label={t("billboard")}
    >
      {/* Billboard Main Layout: Clean split with typography and cutout photography */}
      <div className="relative px-6 py-8 sm:px-10 sm:py-10 md:px-14 md:py-12 flex flex-col md:flex-row items-center justify-between gap-6 md:gap-12 min-h-[290px] sm:min-h-[320px]">
        {/* Left Column: Direct editorial copy */}
        <div className="flex-1 max-w-lg text-left rtl:text-right z-10">
          {/* Promotional badge */}
          <div className="inline-flex items-center gap-2 mb-3">
            <span
              className={`px-2.5 py-0.5 rounded text-[11px] font-black tracking-wider uppercase ${activeSlide.badgeBg}`}
            >
              {activeSlide.badgeText}
            </span>
            <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wide">
              {t(activeSlide.tagKey)}
            </span>
          </div>

          {/* High-contrast solid headline */}
          <h1 className="text-2xl sm:text-3xl md:text-4xl lg:text-[42px] font-black text-foreground tracking-tight leading-tight mb-2.5">
            <span>{t(activeSlide.titleKey)} </span>
            <span className="block text-foreground/80 font-extrabold text-xl sm:text-2xl md:text-3xl mt-0.5">
              {t(activeSlide.highlightKey)}
            </span>
          </h1>

          {/* Clean concise description */}
          <p className="text-muted-foreground text-xs sm:text-sm leading-relaxed mb-6 max-w-md line-clamp-2">
            {t(activeSlide.descKey)}
          </p>

          {/* Clean pill CTA */}
          <div className="flex items-center gap-3">
            <Link
              href={activeSlide.link}
              className={`inline-flex items-center gap-2 px-6 py-2.5 rounded-full text-xs sm:text-sm font-bold shadow-xs transition-transform hover:-translate-y-0.5 cursor-pointer ${activeSlide.ctaClass}`}
            >
              <span>{t(activeSlide.ctaKey)}</span>
              <ArrowRight className="size-4 rtl:rotate-180" />
            </Link>

            <Link
              href="/products"
              className="text-xs sm:text-sm font-semibold text-foreground hover:underline underline-offset-4 px-2 py-1"
            >
              {t("viewDeals")}
            </Link>
          </div>
        </div>

        {/* Right Column: Genuine crisp product photography in clean framing */}
        <div className="w-full md:w-auto shrink-0 flex items-center justify-center z-10">
          <div className="relative size-56 sm:size-64 md:size-72 rounded-2xl overflow-hidden bg-background border border-border/50 shadow-md group-hover/hero:shadow-lg transition-all duration-300 flex items-center justify-center p-3">
            <img
              key={activeSlide.id}
              src={activeSlide.productImage}
              alt={activeSlide.productAlt}
              className="w-full h-full object-cover rounded-xl transition-transform duration-500 ease-out group-hover/hero:scale-102"
            />
          </div>
        </div>
      </div>

      {/* Bottom controls: Clean dot indicators and pause/arrow buttons */}
      <div className="relative px-6 py-2.5 bg-background/60 dark:bg-background/40 backdrop-blur-xs border-t border-border/40 flex items-center justify-between gap-4">
        {/* Pagination Dots */}
        <div className="flex items-center gap-2">
          {SLIDES.map((slide, idx) => {
            const isActive = idx === currentIndex;
            return (
              <button
                key={slide.id}
                onClick={() => setCurrentIndex(idx)}
                aria-label={t("goToSlide", { slide: idx + 1 })}
                className={`h-2.5 rounded-full transition-all duration-200 cursor-pointer ${
                  isActive
                    ? "w-7 bg-foreground"
                    : "w-2 bg-muted-foreground/30 hover:bg-muted-foreground/60"
                } px-1 py-1`}
              />
            );
          })}
        </div>

        {/* Playback & Navigation Controls */}
        <div className="flex items-center gap-1.5">
          <button
            onClick={() => setIsPlaying((prev) => !prev)}
            aria-label={isPlaying ? t("pauseCarousel") : t("playCarousel")}
            className="h-11 w-11 rounded-full bg-background border border-border hover:bg-secondary flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
          >
            {isPlaying ? <Pause className="size-4" /> : <Play className="size-4 ml-0.5" />}
          </button>

          <button
            onClick={prevSlide}
            aria-label={t("previousSlide")}
            className="h-11 w-11 rounded-full bg-background border border-border hover:bg-secondary flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
          >
            <ChevronLeft className="size-4.5 rtl:rotate-180" />
          </button>

          <button
            onClick={nextSlide}
            aria-label={t("nextSlide")}
            className="h-11 w-11 rounded-full bg-background border border-border hover:bg-secondary flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
          >
            <ChevronRight className="size-4.5 rtl:rotate-180" />
          </button>
        </div>
      </div>
    </div>
  );
}
