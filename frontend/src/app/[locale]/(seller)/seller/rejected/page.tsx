"use client";

import { Link } from "@/shared/i18n/navigation";
import {
  SellerSegmentGuard,
  SellerStatusShell,
} from "@/modules/seller";
import { useTranslations } from "next-intl";

export default function SellerRejectedPage() {
  const t = useTranslations("seller.statusPages");

  return (
    <SellerSegmentGuard allow={["rejected"]}>
      <SellerStatusShell
        kind="rejected"
        action={
          <Link
            href="/seller/register"
            className="inline-flex h-9 items-center rounded-md bg-accent-brand px-3 text-sm font-medium text-white hover:bg-accent-brand/90"
          >
            {t("rejected.updateApplication")}
          </Link>
        }
      />
    </SellerSegmentGuard>
  );
}
