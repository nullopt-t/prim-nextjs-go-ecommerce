"use client";

import { useState, useMemo, useRef, useEffect } from "react";
import { useLocale } from "next-intl";
import { Search, ChevronDown, Check, Globe } from "lucide-react";
import { getLocalizedCountries, CountryOption } from "@/utils/countries";

interface CountrySelectProps {
  value: string;
  onChange: (countryName: string, countryCode?: string) => void;
  disabled?: boolean;
  className?: string;
}

export function CountrySelect({
  value,
  onChange,
  disabled = false,
  className = "",
}: CountrySelectProps) {
  const locale = useLocale();
  const [isOpen, setIsOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const containerRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const countries = useMemo(() => getLocalizedCountries(locale), [locale]);

  // Find currently selected country by name or code
  const selectedCountry = useMemo(() => {
    if (!value) return null;
    const lower = value.toLowerCase().trim();
    return (
      countries.find(
        (c) =>
          c.name.toLowerCase() === lower ||
          c.nativeName.toLowerCase() === lower ||
          c.code.toLowerCase() === lower
      ) || null
    );
  }, [value, countries]);

  // Filter countries matching user search query
  const filteredCountries = useMemo(() => {
    if (!searchQuery.trim()) return countries;
    const q = searchQuery.toLowerCase().trim();
    return countries.filter(
      (c) =>
        c.name.toLowerCase().includes(q) ||
        c.nativeName.toLowerCase().includes(q) ||
        c.code.toLowerCase().includes(q)
    );
  }, [searchQuery, countries]);

  // Handle outside clicks
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
        setSearchQuery("");
      }
    }
    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isOpen]);

  // Focus search input when dropdown opens
  useEffect(() => {
    if (isOpen && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isOpen]);

  const handleSelect = (country: CountryOption) => {
    onChange(country.name, country.code);
    setIsOpen(false);
    setSearchQuery("");
  };

  const displayText = selectedCountry ? selectedCountry.name : value || (locale === "ar" ? "اختر الدولة..." : "Select Country...");

  return (
    <div ref={containerRef} className={`relative ${className}`}>
      {/* Trigger Button */}
      <button
        type="button"
        disabled={disabled}
        onClick={() => setIsOpen((prev) => !prev)}
        className="w-full flex items-center justify-between px-3.5 py-2.5 rounded-xl bg-background border border-border/80 text-foreground text-sm hover:border-accent-brand focus:border-accent-brand focus:outline-hidden transition-all text-start cursor-pointer disabled:opacity-50 shadow-2xs"
      >
        <span className="flex items-center gap-2 truncate">
          <Globe className="size-4 text-muted-foreground shrink-0" />
          <span className={selectedCountry || value ? "text-foreground font-medium" : "text-muted-foreground"}>
            {displayText}
          </span>
        </span>
        <ChevronDown
          className={`size-4 text-muted-foreground shrink-0 transition-transform duration-200 ${
            isOpen ? "rotate-180 text-accent-brand" : ""
          }`}
        />
      </button>

      {/* Searchable Dropdown Popup */}
      {isOpen && (
        <div className="absolute z-50 mt-1.5 w-full rounded-2xl bg-card border border-border shadow-xl overflow-hidden animate-in fade-in zoom-in-95 duration-150">
          {/* Search Field */}
          <div className="p-2 border-b border-border/70 flex items-center gap-2 bg-secondary/30">
            <Search className="size-4 text-muted-foreground shrink-0 ml-1.5 rtl:mr-1.5 rtl:ml-0" />
            <input
              ref={inputRef}
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder={locale === "ar" ? "ابحث عن الدولة..." : "Search country..."}
              className="w-full bg-transparent text-foreground text-sm py-1.5 focus:outline-hidden placeholder:text-muted-foreground/70"
            />
          </div>

          {/* Results List */}
          <div className="max-h-60 overflow-y-auto p-1.5 space-y-0.5">
            {filteredCountries.length === 0 ? (
              <div className="py-6 text-center text-xs text-muted-foreground">
                {locale === "ar" ? "لا توجد نتائج مطابقة" : "No countries found"}
              </div>
            ) : (
              filteredCountries.map((country) => {
                const isSelected =
                  selectedCountry?.code === country.code ||
                  value.toLowerCase() === country.name.toLowerCase() ||
                  value.toLowerCase() === country.nativeName.toLowerCase();

                return (
                  <button
                    key={country.code}
                    type="button"
                    onClick={() => handleSelect(country)}
                    className={`w-full flex items-center justify-between px-3 py-2 rounded-xl text-xs sm:text-sm font-medium transition-colors text-start cursor-pointer ${
                      isSelected
                        ? "bg-accent-brand/10 text-accent-brand font-semibold"
                        : "text-foreground hover:bg-secondary/80"
                    }`}
                  >
                    <div className="flex items-center gap-2.5 truncate">
                      <span className="font-mono text-[11px] text-muted-foreground/70 uppercase">
                        {country.code}
                      </span>
                      <span className="truncate">{country.name}</span>
                      {locale === "ar" && country.name !== country.nativeName && (
                        <span className="text-[11px] text-muted-foreground/60 hidden sm:inline truncate">
                          ({country.nativeName})
                        </span>
                      )}
                    </div>
                    {isSelected && <Check className="size-4 text-accent-brand shrink-0" />}
                  </button>
                );
              })
            )}
          </div>
        </div>
      )}
    </div>
  );
}
