"use client";

import { usePathname, Link } from "@/i18n/navigation";
import { useTranslations, useLocale } from "next-intl";
import { ArrowLeft, ArrowRight } from "lucide-react";

export default function OrderDetailsTopNav() {
  const pathname = usePathname();
  const tNav = useTranslations("userNav");
  const tOrders = useTranslations("orders");
  const locale = useLocale();
  const isRtl = locale === "ar";
  const BackIcon = isRtl ? ArrowRight : ArrowLeft;

  // Extract order ID from pathname /user/orders/[id]
  const segments = pathname.split("/");
  const orderId = segments[segments.length - 1];
  const shortId = orderId ? orderId.slice(0, 8).toUpperCase() : "";

  return (
    <nav
      aria-label="Order navigation"
      className="flex items-center gap-2 lg:gap-3.5 overflow-x-auto scrollbar-none px-2"
    >
      <Link
        href="/user/orders"
        className="inline-flex items-center gap-1.5 text-xs lg:text-sm font-medium text-muted-foreground hover:text-foreground py-1 px-1.5 transition-colors whitespace-nowrap"
      >
        <BackIcon className="size-3.5" />
        <span>{tNav("orders")}</span>
      </Link>

      <span className="text-border select-none">/</span>

      <span className="text-xs lg:text-sm font-bold text-foreground py-1 px-1.5 whitespace-nowrap">
        #{shortId || tOrders("title")}
      </span>
    </nav>
  );
}
