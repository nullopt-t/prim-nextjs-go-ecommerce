"use client";

import { useState } from "react";
import { Link, usePathname } from "@/shared/i18n/navigation";
import {
  LayoutDashboard,
  Package,
  Boxes,
  ShoppingBag,
  BadgePercent,
  PackagePlus,
  DollarSign,
  Settings2,
  Menu,
  X,
} from "lucide-react";
import { useTranslations } from "next-intl";
import { useSellerContext } from "./sellerContext";

export function SellerLayout({ children }: { children?: React.ReactNode }) {
  const t = useTranslations("seller");
  const pathname = usePathname();
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const { seller } = useSellerContext();

  const navGroups = [
    {
      key: "overview",
      items: [
        { label: t("layout.nav.dashboard"), path: "/seller/dashboard", icon: LayoutDashboard },
      ],
    },
    {
      key: "catalog",
      items: [
        { label: t("layout.nav.products"), path: "/seller/products", icon: Package },
        { label: t("layout.nav.bundles"), path: "/seller/bundles", icon: PackagePlus },
        { label: t("layout.nav.inventory"), path: "/seller/inventory", icon: Boxes },
      ],
    },
    {
      key: "sales",
      items: [
        { label: t("layout.nav.orders"), path: "/seller/orders", icon: ShoppingBag },
        { label: t("layout.nav.promotions"), path: "/seller/promotions", icon: BadgePercent },
        { label: t("layout.nav.payouts"), path: "/seller/payouts", icon: DollarSign },
      ],
    },
    {
      key: "account",
      items: [
        { label: t("layout.nav.settings"), path: "/seller/settings", icon: Settings2 },
      ],
    },
  ];

  if (!seller) return null;

  return (
    <div className="flex min-h-screen flex-col bg-muted/20 font-sans text-foreground">
      <header className="sticky top-0 z-40 flex h-14 items-center justify-between border-b border-border bg-background px-4 lg:px-6">
        <div className="flex items-center gap-3">
          <button
            onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
            className="p-2 text-muted-foreground hover:text-foreground lg:hidden"
            aria-label={t("layout.menu")}
          >
            {isMobileMenuOpen ? <X className="size-5" /> : <Menu className="size-5" />}
          </button>
          <Link href="/" className="flex items-center gap-2 hover:opacity-90">
            <span dir="ltr" className="text-lg font-black tracking-tighter text-foreground">
              PRI<span className="text-accent-brand">M</span>
            </span>
            <span className="hidden border-s border-border ps-3 text-xs font-medium text-muted-foreground sm:inline">
              {t("layout.sellerHub")}
            </span>
          </Link>
        </div>

        <div className="flex items-center gap-2">
          <span className="hidden max-w-48 truncate text-xs text-muted-foreground sm:inline">
            {seller.storeName}
          </span>
          <Link
            href="/portal"
            className="hidden rounded-md px-2.5 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground md:inline-flex"
          >
            {t("layout.routeManager")}
          </Link>
          <Link
            href="/"
            className="rounded-md border border-border px-2.5 py-1.5 text-xs font-medium text-foreground transition-colors hover:bg-secondary"
          >
            {t("layout.storefront")}
          </Link>
        </div>
      </header>

      <div className="relative flex flex-1">
        {isMobileMenuOpen ? (
          <div
            onClick={() => setIsMobileMenuOpen(false)}
            className="fixed inset-0 z-30 bg-black/40 lg:hidden"
          />
        ) : null}

        <aside
          className={`fixed start-0 top-14 z-40 flex h-[calc(100dvh-3.5rem)] w-60 flex-col border-e border-border bg-sidebar transition-transform duration-200 ease-out lg:sticky ${
            isMobileMenuOpen ? "translate-x-0" : "max-lg:-translate-x-full max-lg:rtl:translate-x-full"
          }`}
        >
          <div className="p-3">
            <div className="rounded-md border border-border bg-background p-3">
              <div className="truncate text-sm font-medium text-foreground">{seller.storeName}</div>
              <div className="mt-0.5 truncate text-xs text-muted-foreground">{seller.slug}</div>
              <div className="mt-2 flex items-center justify-between text-xs text-muted-foreground">
                <span>{t("layout.commissionRate")}</span>
                <span className="font-medium tabular-nums text-foreground">
                  {(seller.commissionRateBps / 100).toFixed(2)}%
                </span>
              </div>
            </div>
          </div>

          <nav className="flex-1 space-y-5 overflow-y-auto px-3 pb-3">
            {navGroups.map((group) => (
              <div key={group.key} className="space-y-0.5">
                <div className="px-3 pb-1 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
                  {t(`layout.groups.${group.key}`)}
                </div>
                {group.items.map((item) => {
                  const Icon = item.icon;
                  const isActive =
                    pathname === item.path ||
                    pathname.startsWith(`${item.path}/`) ||
                    (item.path === "/seller/dashboard" && pathname === "/seller");
                  return (
                    <Link
                      key={item.path}
                      href={item.path}
                      onClick={() => setIsMobileMenuOpen(false)}
                      aria-current={isActive ? "page" : undefined}
                      className={`flex items-center gap-2.5 rounded-md px-3 py-2 text-sm transition-colors ${
                        isActive
                          ? "bg-background font-medium text-foreground shadow-xs"
                          : "text-muted-foreground hover:bg-background/70 hover:text-foreground"
                      }`}
                    >
                      <Icon className="size-4 shrink-0" />
                      <span className="truncate">{item.label}</span>
                    </Link>
                  );
                })}
              </div>
            ))}
          </nav>

          <div className="border-t border-border p-3">
            <div className="text-xs font-medium text-foreground">{t("layout.support.title")}</div>
            <a
              href="mailto:sellers@marketplace.com"
              className="mt-1 block text-xs text-muted-foreground hover:text-foreground hover:underline"
            >
              sellers@marketplace.com
            </a>
          </div>
        </aside>

        <main className="min-w-0 flex-1 overflow-y-auto p-4 md:p-6">
          <div className="mx-auto w-full max-w-[1400px]">{children}</div>
        </main>
      </div>
    </div>
  );
}
