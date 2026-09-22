"use client";

import { Search, Mic } from "lucide-react";
import { useTranslations } from "next-intl";

export function SearchBar() {
  const t = useTranslations("common");

  return (
    <div className="flex items-center relative bg-input-background rounded-lg border-2 border-border group focus-within:border-accent-brand px-3">
      <Search className="text-muted-foreground size-5 group-focus-within:text-accent-brand shrink-0" />
      <input
        type="text"
        className="w-full text-txt-sm md:text-txt-md lg:text-txt-lg text-foreground bg-transparent placeholder:text-muted-foreground p-2 outline-none"
        placeholder={t("header.search")}
      />
      <Mic className="text-muted-foreground cursor-pointer hover:text-accent-brand size-5 shrink-0" />
    </div>
  );
}
