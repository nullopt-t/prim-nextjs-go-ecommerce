"use client";

import { useEffect } from "react";
import { useRouter } from "@/shared/i18n/navigation";
import { ErrorState } from "@/shared/ui/errorState";
import { useTranslations } from "next-intl";
import { useSellerContext } from "./sellerContext";
import { SellerLayout } from "./sellerLayout";

export function SellerConsoleGate({ children }: { children: React.ReactNode }) {
  const t = useTranslations("seller");
  const { state, loading, loadError, refreshSeller } = useSellerContext();
  const router = useRouter();

  useEffect(() => {
    if (!loading && state !== "active" && state !== "error") {
      router.replace("/seller");
    }
  }, [loading, router, state]);

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

  if (loading || state !== "active") {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-accent-brand border-t-transparent" />
      </div>
    );
  }

  return <SellerLayout>{children}</SellerLayout>;
}
