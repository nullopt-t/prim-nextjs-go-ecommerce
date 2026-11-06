"use client";

import { useEffect } from "react";
import { useRouter } from "@/shared/i18n/navigation";
import { useVendorContext, type VendorState } from "./vendorContext";

export function VendorSegmentGuard({
  allow,
  children,
}: {
  allow: VendorState[];
  children: React.ReactNode;
}) {
  const { state, loading } = useVendorContext();
  const router = useRouter();
  const allowed = allow.includes(state);

  useEffect(() => {
    if (!loading && !allowed) {
      router.replace("/vendor");
    }
  }, [allowed, loading, router]);

  if (loading || !allowed) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-accent-brand border-t-transparent" />
      </div>
    );
  }

  return <>{children}</>;
}
