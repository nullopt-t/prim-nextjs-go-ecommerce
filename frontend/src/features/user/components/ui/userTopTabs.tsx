"use client";

import { usePathname, Link } from "@/i18n/navigation";
import { useTranslations } from "next-intl";

export const accountTabKeys = [
  { id: "overview", key: "overview", path: "/user/overview" },
  { id: "orders", key: "orders", path: "/user/orders" },
  { id: "address", key: "addresses", path: "/user/address" },
  { id: "payment", key: "payment", path: "/user/payment" },
  { id: "reviews", key: "reviews", path: "/user/reviews" },
  { id: "settings", key: "settings", path: "/user/settings" },
] as const;

export default function UserTopTabs() {
  const pathname = usePathname();
  const t = useTranslations("userNav");

  return (
    <nav
      aria-label="Account navigation"
      className="flex items-center gap-1 sm:gap-2 lg:gap-3.5 overflow-x-auto scrollbar-none px-2"
    >
      {accountTabKeys.map((tab) => {
        const isActive = pathname === tab.path || pathname.startsWith(`${tab.path}/`);

        return (
          <Link
            key={tab.id}
            href={tab.path}
            className={`py-1 px-1.5 text-xs lg:text-sm font-medium transition-colors whitespace-nowrap ${
              isActive
                ? "text-foreground font-bold"
                : "text-muted-foreground hover:text-foreground"
            }`}
          >
            {t(tab.key)}
          </Link>
        );
      })}
    </nav>
  );
}
