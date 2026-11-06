"use client";

import { useEffect } from "react";
import { useRouter } from "@/shared/i18n/navigation";
import { ErrorState } from "@/shared/ui/errorState";
import { useTranslations } from "next-intl";
import { useSellerContext } from "./sellerContext";

const DESTINATIONS = {
  unregistered: "/seller/register",
  pending: "/seller/pending",
  active: "/seller/dashboard",
  suspended: "/seller/suspended",
  rejected: "/seller/rejected",
} as const;

export function SellerResolver() {
  const t = useTranslations("seller");
  const { state, refreshSeller, loadError } = useSellerContext();
  const router = useRouter();

  useEffect(() => {
    if (state === "loading" || state === "error" || state === "unauthenticated") return;
    router.replace(DESTINATIONS[state]);
  }, [router, state]);

  if (state === "error") {
    return (
      <div className="mx-auto w-full max-w-xl px-4 py-16">
        <ErrorState
          title={t("loadFailed")}
          description={loadError?.message}
          onRetry={() => void refreshSeller()}
        />
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="size-8 animate-spin rounded-full border-2 border-accent-brand border-t-transparent" />
    </div>
  );
}
