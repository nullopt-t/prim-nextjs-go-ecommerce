"use client";

import { useMemo, useState } from "react";
import { Link } from "@/shared/i18n/navigation";
import {
  LayoutDashboard,
  Package,
  Boxes,
  ShoppingBag,
  DollarSign,
  Store,
  ShieldCheck,
  Menu,
  X,
} from "lucide-react";
import { useTranslations } from "next-intl";
import { ErrorState } from "@/shared/ui/errorState";
import { VendorProvider, useVendorContext } from "./vendorContext";

function VendorLayoutInner({ children }: { children?: React.ReactNode }) {
  const t = useTranslations("vendor");
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const { vendor, loading, loadError, refreshVendor } = useVendorContext();

  const navItems = [
    { label: t("layout.nav.dashboard"), path: "/vendor/dashboard", icon: LayoutDashboard },
    { label: t("layout.nav.products"), path: "/vendor/products", icon: Package },
    { label: t("layout.nav.inventory"), path: "/vendor/inventory", icon: Boxes },
    { label: t("layout.nav.orders"), path: "/vendor/orders", icon: ShoppingBag },
    { label: t("layout.nav.payouts"), path: "/vendor/payouts", icon: DollarSign },
  ];

  const statusMessage = useMemo(() => {
    if (!vendor) return "";
    if (vendor.status === "pending") return t("statusPending");
    if (vendor.status === "suspended") return t("statusSuspended");
    if (vendor.status === "rejected") return t("statusRejected");
    return "";
  }, [t, vendor]);

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-accent-brand border-t-transparent" />
      </div>
    );
  }

  if (loadError || !vendor) {
    return (
      <div className="mx-auto max-w-3xl p-4 md:p-8">
        <ErrorState
          title={t("profileLoadTitle")}
          description={loadError?.message || t("profileLoadDescription")}
          onRetry={() => void refreshVendor()}
        />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-secondary/15 text-foreground flex flex-col font-sans">
      <header className="h-16 bg-background border-b border-border sticky top-0 z-40 px-4 lg:px-6 flex items-center justify-between shadow-xs">
        <div className="flex items-center gap-4">
          <button
            onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
            className="lg:hidden p-2 text-muted-foreground hover:text-foreground"
            aria-label={t("layout.menu")}
          >
            {isMobileMenuOpen ? <X className="size-5" /> : <Menu className="size-5" />}
          </button>
          <Link href="/" className="flex items-center gap-2 group hover:opacity-90 transition-opacity">
            <span dir="ltr" className="font-black text-xl tracking-tighter text-foreground">
              PRI<span className="text-accent-brand">M</span>
            </span>
            <div>
              <span className="font-bold text-xs tracking-tight text-foreground block leading-none">
                {t("layout.vendorHub")}
              </span>
              <span className="text-[9px] text-muted-foreground">{t("layout.merchant")}</span>
            </div>
          </Link>
        </div>

        <div className="flex items-center gap-3">
          <span className="hidden sm:inline-flex rounded-full border border-border bg-secondary/40 px-3 py-1 text-xs font-semibold text-muted-foreground">
            {vendor.storeName}
          </span>
          <Link
            href="/portal"
            className="hidden md:flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-xl bg-accent-brand/10 text-accent-brand border border-accent-brand/20 hover:bg-accent-brand/20 transition-colors"
          >
            <ShieldCheck className="size-3.5" />
            <span>{t("layout.routeManager")}</span>
          </Link>
          <Link
            href="/"
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-xl bg-secondary text-foreground border border-border hover:bg-secondary/80 transition-colors"
          >
            <Store className="size-3.5 text-accent-brand" />
            <span className="hidden sm:inline">{t("layout.storefront")}</span>
          </Link>
        </div>
      </header>

      <div className="flex-1 flex overflow-hidden">
        <aside
          className={`${
            isMobileMenuOpen ? "block fixed inset-0 z-50 top-16 bg-background" : "hidden"
          } lg:block lg:w-64 border-r border-border bg-background/50 backdrop-blur-xs p-4 shrink-0`}
        >
          <div className="mb-6 p-3 bg-accent-brand/5 border border-accent-brand/20 rounded-xl">
            <div className="text-[10px] font-bold text-accent-brand uppercase tracking-wider">
              {t("layout.activeAccount")}
            </div>
            <div className="font-bold text-sm text-foreground mt-0.5">{vendor.storeName}</div>
            <div className="text-xs text-muted-foreground">{vendor.slug}</div>
            <div className="mt-2 text-[11px] font-medium text-emerald-600 dark:text-emerald-400 flex items-center justify-between">
              <span>{t("layout.commissionRate")}</span>
              <span className="font-bold">{(vendor.commissionRateBps / 100).toFixed(2)}%</span>
            </div>
            {statusMessage && (
              <div className="mt-2 text-[11px] font-semibold text-amber-700 dark:text-amber-400">
                {statusMessage}
              </div>
            )}
          </div>

          <nav className="space-y-1">
            {navItems.map((item) => {
              const Icon = item.icon;
              return (
                <Link
                  key={item.path}
                  href={item.path}
                  onClick={() => setIsMobileMenuOpen(false)}
                  className="flex items-center gap-3 px-3.5 py-2.5 rounded-xl text-xs font-semibold text-muted-foreground hover:bg-secondary hover:text-foreground transition-all"
                >
                  <Icon className="size-4 shrink-0" />
                  <span>{item.label}</span>
                </Link>
              );
            })}
          </nav>

          <div className="mt-8 p-4 bg-secondary/30 rounded-2xl border border-border text-xs space-y-2">
            <div className="font-bold text-foreground">{t("layout.support.title")}</div>
            <p className="text-muted-foreground text-[11px] leading-relaxed">
              {t("layout.support.description")}
            </p>
            <a
              href="mailto:vendors@marketplace.com"
              className="text-accent-brand font-semibold text-[11px] hover:underline block"
            >
              vendors@marketplace.com
            </a>
          </div>
        </aside>

        <main className="flex-1 overflow-y-auto p-4 md:p-8">{children}</main>
      </div>
    </div>
  );
}

export function VendorLayout({ children }: { children?: React.ReactNode }) {
  return (
    <VendorProvider>
      <VendorLayoutInner>{children}</VendorLayoutInner>
    </VendorProvider>
  );
}
