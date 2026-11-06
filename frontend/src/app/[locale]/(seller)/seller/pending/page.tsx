"use client";

import {
  SellerSegmentGuard,
  SellerStatusShell,
} from "@/modules/seller";

export default function SellerPendingPage() {
  return (
    <SellerSegmentGuard allow={["pending"]}>
      <SellerStatusShell kind="pending" />
    </SellerSegmentGuard>
  );
}
