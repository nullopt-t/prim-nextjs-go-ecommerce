"use client";

import { usePathname, Link } from "@/i18n/navigation";
import { useTranslations } from "next-intl";
import { useAuthContext } from "@/context/AuthContext";
import { Shield } from "lucide-react";

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
  const { user } = useAuthContext();
  const isStaff = user?.role === "admin" || user?.role === "superadmin";

  return (
    <nav
      aria-label={t("accountNav")}
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

      {isStaff && (
        <Link
          href="/admin/customers"
          className="ml-auto py-1 px-2.5 rounded-lg bg-purple-500/10 text-purple-700 dark:text-purple-300 border border-purple-500/20 text-xs font-bold transition-all hover:bg-purple-500/20 flex items-center gap-1.5 whitespace-nowrap"
        >
          <Shield className="size-3.5" />
          <span>{t("adminConsole")}</span>
        </Link>
      )}
    </nav>
  );
}
