"use client";

import { usePathname, Link } from "@/i18n/navigation";
import { Blocks, Package, Heart, Settings } from "lucide-react";
const mobileTabs = [
  { id: "overview", icon: Blocks, label: "Overview", path: "/user/overview" },
  { id: "orders", icon: Package, label: "Orders", path: "/user/orders" },
  { id: "wishlist", icon: Heart, label: "Wishlist", path: "/user/wishlist" },
  { id: "settings", icon: Settings, label: "Settings", path: "/user/settings" },
];

export default function MobileNav() {
  const pathname = usePathname();

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
              <span className="truncate leading-none">{item.label}</span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
