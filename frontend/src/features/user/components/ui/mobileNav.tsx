"use client";

import { usePathname, Link } from "@/i18n/navigation";
import { useTranslations } from "next-intl";
import { User, Settings, Store } from "lucide-react";

export const mobileNavTabs = [
  { id: "store", icon: Store, key: "store", path: "/" },
  {
    id: "you",
    icon: User,
    key: "overview",
    path: "/user/overview",
    matchPrefixes: ["/user/overview", "/user/orders", "/user/address", "/user/wishlist", "/user/payment", "/user/reviews"],
  },
  {
    id: "settings",
    icon: Settings,
    key: "settings",
    path: "/user/settings",
    matchPrefixes: ["/user/settings"],
  },
] as const;

export default function MobileNav() {
  const pathname = usePathname();
  const t = useTranslations("userNav");

  return (
    <nav
      aria-label={t("mobileNav")}
      className="md:hidden fixed bottom-0 inset-x-0 z-50 bg-card/95 backdrop-blur-md border-t border-border shadow-lg pb-safe"
    >
      <div className="grid grid-cols-3 w-full items-center px-3 py-2">
        {mobileNavTabs.map((item) => {
          const isActive =
            item.id === "store"
              ? pathname === "/"
              : item.matchPrefixes.some(
                  (prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`)
                );

          const Icon = item.icon;

          return (
            <Link
              key={item.id}
              href={item.path}
              className={`flex flex-col items-center justify-center gap-1 py-1 rounded-xl text-xs transition-all cursor-pointer ${
                isActive
                  ? "text-accent-brand font-semibold"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              <div
                className={`flex items-center justify-center size-8 rounded-full transition-colors ${
                  isActive ? "bg-accent-brand/10 text-accent-brand" : ""
                }`}
              >
                <Icon className="size-5" />
              </div>
              <span className="truncate leading-none text-[11px]">{t(item.key)}</span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
