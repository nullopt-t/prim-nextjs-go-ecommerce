"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { Link, usePathname } from "@/shared/i18n/navigation";
import { Menu, X } from "lucide-react";

const NAV_LINKS = [
  { id: "nav-home", path: "/", label: "header.home" },
  { id: "nav-products", path: "/products", label: "header.products" },
  { id: "nav-categories", path: "/categories", label: "header.categories" },
  { id: "nav-sell", path: "/sell", label: "header.sell" },
] as const;

function useIsActive() {
  const pathname = usePathname();
  return (path: string) =>
    path === "/" ? pathname === "/" : pathname === path || pathname.startsWith(`${path}/`);
}

/** Desktop primary navigation row for the storefront header. */
export function DesktopNav() {
  const t = useTranslations("common");
  const isActive = useIsActive();

  return (
    <nav aria-label={t("header.nav")} className="hidden md:block">
      <ul className="flex items-center gap-6">
        {NAV_LINKS.map((item) => (
          <li key={item.id} className="cursor-pointer whitespace-nowrap">
            <Link
              href={item.path}
              aria-current={isActive(item.path) ? "page" : undefined}
              className={`transition-colors py-1.5 text-sm font-medium ${
                isActive(item.path)
                  ? "text-accent-brand underline underline-offset-8 decoration-2 decoration-accent-brand"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              {t(item.label)}
            </Link>
          </li>
        ))}
      </ul>
    </nav>
  );
}

/** Mobile primary navigation: menu button with a dropdown panel. */
export function MobileNav() {
  const t = useTranslations("common");
  const isActive = useIsActive();
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <nav aria-label={t("header.nav")} className="relative md:hidden">
      <button
        type="button"
        onClick={() => setMobileOpen((prev) => !prev)}
        aria-expanded={mobileOpen}
        aria-label={t("header.menu")}
        className="flex size-9 items-center justify-center rounded-xl border border-border/40 text-muted-foreground transition-colors hover:bg-secondary/70 hover:text-foreground"
      >
        {mobileOpen ? <X className="size-5" /> : <Menu className="size-5" />}
      </button>
      {mobileOpen ? (
        <ul className="absolute start-0 top-full z-50 mt-2 w-48 rounded-lg border border-border bg-background p-2 shadow-lg">
          {NAV_LINKS.map((item) => (
            <li key={item.id}>
              <Link
                href={item.path}
                onClick={() => setMobileOpen(false)}
                aria-current={isActive(item.path) ? "page" : undefined}
                className={`block rounded-md px-3 py-2 text-sm font-medium ${
                  isActive(item.path)
                    ? "text-accent-brand"
                    : "text-muted-foreground hover:text-foreground"
                }`}
              >
                {t(item.label)}
              </Link>
            </li>
          ))}
        </ul>
      ) : null}
    </nav>
  );
}
