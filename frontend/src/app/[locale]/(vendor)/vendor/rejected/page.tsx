"use client";

import { Link } from "@/shared/i18n/navigation";
import {
  VendorSegmentGuard,
  VendorStatusShell,
} from "@/modules/vendor";
import { useTranslations } from "next-intl";

export default function VendorRejectedPage() {
  const t = useTranslations("vendor.statusPages");

  return (
    <VendorSegmentGuard allow={["rejected"]}>
      <VendorStatusShell
        kind="rejected"
        action={
          <Link
            href="/vendor/register"
            className="inline-flex h-9 items-center rounded-md bg-accent-brand px-3 text-sm font-medium text-white hover:bg-accent-brand/90"
          >
            {t("rejected.updateApplication")}
          </Link>
        }
      />
    </VendorSegmentGuard>
  );
}
