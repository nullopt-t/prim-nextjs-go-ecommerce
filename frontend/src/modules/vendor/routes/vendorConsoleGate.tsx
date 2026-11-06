"use client";

import { useEffect } from "react";
import { useRouter } from "@/shared/i18n/navigation";
import { ErrorState } from "@/shared/ui/errorState";
import { useTranslations } from "next-intl";
import { useVendorContext } from "./vendorContext";
import { VendorLayout } from "./vendorLayout";

export function VendorConsoleGate({ children }: { children: React.ReactNode }) {
  const t = useTranslations("vendor");
  const { state, loading, loadError, refreshVendor } = useVendorContext();
  const router = useRouter();

  useEffect(() => {
    if (!loading && state !== "active" && state !== "error") {
      router.replace("/vendor");
    }
  }, [loading, router, state]);

  if (state === "error") {
    return (
      <div className="mx-auto w-full max-w-xl px-4 py-16">
        <ErrorState
          title={t("loadFailed")}
          description={loadError?.message}
          onRetry={() => void refreshVendor()}
        />
      </div>
    );
  }

  if (loading || state !== "active") {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-accent-brand border-t-transparent" />
      </div>
    );
  }

  return <VendorLayout>{children}</VendorLayout>;
}
