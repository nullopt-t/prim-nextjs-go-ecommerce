"use client";

import { Search, X } from "lucide-react";
import { useTranslations } from "next-intl";
import { useRouter } from "@/i18n/navigation";
import { useSearchParams } from "next/navigation";
import { useState, useEffect } from "react";

export function SearchBar({ className = "" }: { className?: string }) {
  const t = useTranslations("common");
  const router = useRouter();
  const searchParams = useSearchParams();
  const [query, setQuery] = useState("");

  useEffect(() => {
    const q = searchParams.get("q") || "";
    setQuery(q);
  }, [searchParams]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = query.trim();
    if (trimmed) {
      router.push(`/products?q=${encodeURIComponent(trimmed)}`);
    } else {
      router.push("/products");
    }
  };

  const handleClear = () => {
    setQuery("");
    const currentQ = searchParams.get("q");
    if (currentQ) {
      router.push("/products");
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      className={`flex items-center relative bg-secondary/40 hover:bg-secondary/60 focus-within:bg-background rounded-full border border-border/80 focus-within:border-accent-brand focus-within:ring-2 focus-within:ring-accent-brand/15 px-3.5 py-1.5 transition-all duration-200 ${className}`}
    >
      <Search className="text-muted-foreground size-4 shrink-0 transition-colors" />
      <input
        type="search"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        className="w-full text-xs sm:text-sm text-foreground bg-transparent placeholder:text-muted-foreground px-2.5 py-1 outline-none font-normal"
        placeholder={t("header.search")}
      />
      {query && (
        <button
          type="button"
          onClick={handleClear}
          aria-label="Clear search"
          className="text-muted-foreground hover:text-foreground p-0.5 rounded-full hover:bg-muted transition-colors cursor-pointer shrink-0"
        >
          <X className="size-3.5" />
        </button>
      )}
    </form>
  );
}
