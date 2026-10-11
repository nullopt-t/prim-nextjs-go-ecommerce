"use client";

import { usePathname, Link } from "@/i18n/navigation";
import { useTranslations, useLocale } from "next-intl";
import { ArrowLeft, ArrowRight } from "lucide-react";
import { settingsSubPages } from "@/features/settings/components/layout/settingsShell";

export default function SettingsTopTabs() {
  const pathname = usePathname();
  const tSettings = useTranslations("settings");
  const tNav = useTranslations("userNav");
  const locale = useLocale();
  const isRtl = locale === "ar";
  const BackIcon = isRtl ? ArrowRight : ArrowLeft;

  return (
    <nav
      aria-label="Settings navigation"
      className="flex items-center gap-1 sm:gap-2 lg:gap-3.5 overflow-x-auto scrollbar-none px-2"
    >
      <Link
        href="/user/overview"
        className="inline-flex items-center gap-1 text-xs lg:text-sm font-medium text-muted-foreground hover:text-foreground py-1 px-1.5 transition-colors whitespace-nowrap"
      >
        <BackIcon className="size-3.5" />
        <span>{tNav("overview")}</span>
      </Link>

      <span className="text-border select-none">|</span>

      {settingsSubPages.map((sub) => {
        const isActive =
          pathname === sub.path ||
          pathname.startsWith(`${sub.path}/`) ||
          (sub.id === "profile" && pathname === "/user/settings");

        return (
          <Link
            key={sub.id}
            href={sub.path}
            className={`py-1 px-1.5 text-xs lg:text-sm font-medium transition-colors whitespace-nowrap ${
              isActive
                ? sub.isDanger
                  ? "text-destructive font-bold"
                  : "text-foreground font-bold"
                : sub.isDanger
                ? "text-muted-foreground hover:text-destructive"
                : "text-muted-foreground hover:text-foreground"
            }`}
          >
            {tSettings(sub.labelKey as any)}
          </Link>
        );
      })}
    </nav>
  );
}
