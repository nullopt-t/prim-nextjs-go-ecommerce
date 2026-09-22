"use client";

import { useTranslations } from "next-intl";
import { Link, usePathname } from "@/i18n/navigation";

export function Categories() {
  const t = useTranslations("common");
  const pathname = usePathname();

  const navLinks = [
    {
      id: "nav-home",
      path: "/",
      label: "header.home",
    },
    {
      id: "nav-products",
      path: "/products",
      label: "header.allCategories",
    },
  ];

  return (
    <ul className="flex items-center gap-6 text-txt-sm md:text-txt-md font-medium">
      {navLinks.map((item) => {
        const isActive = item.path === "/" ? pathname === "/" : pathname.startsWith(item.path);
        return (
          <li key={item.id} className="cursor-pointer group whitespace-nowrap">
            <Link
              href={item.path}
              className={`transition-colors py-1 ${
                isActive
                  ? "text-accent-brand font-semibold underline underline-offset-8 decoration-2 decoration-accent-brand"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              {t(item.label)}
            </Link>
          </li>
        );
      })}
    </ul>
  );
}
