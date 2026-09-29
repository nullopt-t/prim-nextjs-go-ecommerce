"use client";

import { useState, useEffect, useRef } from "react";
import { Search, X, ArrowRight, ArrowLeft } from "lucide-react";
import { useTranslations, useLocale } from "next-intl";
import { useRouter } from "@/i18n/navigation";
import { useSearchParams } from "next/navigation";
import { useCatalogContext } from "@/context/CatalogContext";

export function MobileSearchDialog() {
  const [isOpen, setIsOpen] = useState(false);
  const [query, setQuery] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);
  const dialogRef = useRef<HTMLDivElement>(null);

  const t = useTranslations("common.header");
  const locale = useLocale();
  const router = useRouter();
  const searchParams = useSearchParams();
  const { categories } = useCatalogContext();

  const isRTL = locale === "ar";
  const ArrowIcon = isRTL ? ArrowLeft : ArrowRight;

  // Sync search param
  useEffect(() => {
    const q = searchParams.get("q") || "";
    setQuery(q);
  }, [searchParams]);

  // Autofocus input when dialog opens & manage body scroll lock
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = "hidden";
      // Small timeout for smooth animation before focus
      const timer = setTimeout(() => {
        inputRef.current?.focus();
      }, 100);
      return () => {
        clearTimeout(timer);
        document.body.style.overflow = "";
      };
    } else {
      document.body.style.overflow = "";
    }
  }, [isOpen]);

  // Handle escape key
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && isOpen) {
        setIsOpen(false);
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = query.trim();
    setIsOpen(false);
    if (trimmed) {
      router.push(`/products?q=${encodeURIComponent(trimmed)}`);
    } else {
      router.push("/products");
    }
  };

  const handleSelectCategory = (catId: string) => {
    setIsOpen(false);
    router.push(`/products?category=${catId}`);
  };

  return (
    <>
      {/* Trigger Button - visible only on mobile/tablet (< md) */}
      <button
        type="button"
        onClick={() => setIsOpen(true)}
        aria-label={t("search")}
        aria-haspopup="dialog"
        className="size-9 sm:size-10 rounded-xl text-muted-foreground hover:text-accent-brand hover:bg-secondary/70 transition-all flex items-center justify-center cursor-pointer border border-border/40 hover:border-accent-brand/40 shadow-2xs md:hidden"
      >
        <Search className="size-4.5 sm:size-5" />
      </button>

      {/* Fullscreen Search Dialog Modal */}
      {isOpen && (
        <div
          role="dialog"
          aria-modal="true"
          aria-label={t("searchProducts")}
          className="fixed inset-0 z-[100] bg-background/95 backdrop-blur-md md:hidden flex flex-col animate-in fade-in-0 duration-200"
        >
          {/* Top Search Bar */}
          <div className="flex items-center gap-2.5 px-4 h-16 border-b border-border/80 bg-background/80">
            <form onSubmit={handleSubmit} className="flex-1 flex items-center relative">
              <Search className="absolute ltr:left-3.5 rtl:right-3.5 size-4.5 text-muted-foreground pointer-events-none" />
              <input
                ref={inputRef}
                type="search"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder={t("searchPlaceholder")}
                className="w-full h-11 ltr:pl-10 ltr:pr-9 rtl:pr-10 rtl:pl-9 rounded-xl bg-secondary/60 text-foreground text-sm placeholder:text-muted-foreground border border-border/60 focus:border-accent-brand focus:ring-2 focus:ring-accent-brand/15 outline-none transition-all"
              />
              {query && (
                <button
                  type="button"
                  onClick={() => setQuery("")}
                  aria-label={t("clear")}
                  className="absolute ltr:right-2.5 rtl:left-2.5 text-muted-foreground hover:text-foreground p-1 rounded-full hover:bg-muted transition-colors cursor-pointer"
                >
                  <X className="size-3.5" />
                </button>
              )}
            </form>

            <button
              type="button"
              onClick={() => setIsOpen(false)}
              className="text-xs font-semibold px-3 py-2 rounded-xl text-muted-foreground hover:text-foreground hover:bg-secondary/80 transition-colors cursor-pointer shrink-0"
            >
              {t("close")}
            </button>
          </div>

          {/* Quick Categories & Suggestions */}
          <div className="flex-1 overflow-y-auto p-4 space-y-4">
            {categories && categories.length > 0 && (
              <div>
                <p className="text-xs font-bold text-muted-foreground uppercase tracking-wider mb-2.5 px-1">
                  {t("popularCategories")}
                </p>
                <div className="flex flex-wrap gap-2">
                  {categories.slice(0, 8).map((cat: any) => (
                    <button
                      key={cat.id || cat.slug || cat.name}
                      type="button"
                      onClick={() => handleSelectCategory(cat.id || cat.slug)}
                      className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-secondary/70 hover:bg-accent-brand hover:text-white text-foreground border border-border/50 transition-all cursor-pointer"
                    >
                      <span>{cat.name}</span>
                      <ArrowIcon className="size-3 opacity-60" />
                    </button>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </>
  );
}
