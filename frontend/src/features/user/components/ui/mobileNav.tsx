"use client";

import { usePathname, Link } from "@/i18n/navigation";
import { useTranslations } from "next-intl";
import { Blocks, Package, Heart, Store } from "lucide-react";

const mobileTabs = [
  { id: "store", icon: Store, key: "store", path: "/" },
  { id: "overview", icon: Blocks, key: "overview", path: "/user/overview" },
  { id: "orders", icon: Package, key: "orders", path: "/user/orders" },
  { id: "wishlist", icon: Heart, key: "wishlist", path: "/user/wishlist" },
] as const;

export default function MobileNav() {
  const pathname = usePathname();
  const t = useTranslations("userNav");

  return (
    <nav className="md:hidden fixed bottom-0 inset-x-0 z-50 bg-card/95 backdrop-blur-sm border-t border-border pb-safe">
      <div className="grid grid-cols-4 items-center">
        {mobileTabs.map((item) => {
          const isActive = pathname === item.path;
          return (
            <Link
              key={item.id}
              href={item.path}
              className={`flex flex-col items-center justify-center gap-1 py-2.5 px-1 text-[11px] font-medium transition-colors ${
                isActive
                  ? "text-accent-brand"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              <item.icon className={`size-5 ${isActive ? "text-accent-brand" : ""}`} />
              <span className="truncate leading-none">{t(item.key)}</span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
