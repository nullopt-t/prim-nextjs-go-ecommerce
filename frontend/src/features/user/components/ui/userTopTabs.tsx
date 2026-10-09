"use client";

import { usePathname, Link } from "@/i18n/navigation";
import { useTranslations } from "next-intl";

export const accountTabKeys = [
  { id: "overview", key: "overview", path: "/user/overview" },
  { id: "orders", key: "orders", path: "/user/orders" },
  { id: "address", key: "addresses", path: "/user/address" },
  { id: "payment", key: "payment", path: "/user/payment" },
  { id: "reviews", key: "reviews", path: "/user/reviews" },
] as const;

export default function UserTopTabs() {
  const pathname = usePathname();
  const t = useTranslations("userNav");

  return (
    <div className="w-full border-b border-border mb-6">
      <nav aria-label="Account navigation" className="flex items-center gap-6 overflow-x-auto scrollbar-none -mb-px">
        {accountTabKeys.map((tab) => {
          const isActive = pathname === tab.path;

          return (
            <Link
              key={tab.id}
              href={tab.path}
              className={`py-3 text-sm font-medium border-b-2 whitespace-nowrap transition-colors duration-150 ${
                isActive
                  ? "border-foreground text-foreground font-semibold"
                  : "border-transparent text-muted-foreground hover:text-foreground hover:border-border"
              }`}
            >
              {t(tab.key)}
            </Link>
          );
        })}
      </nav>
    </div>
  );
}
