"use client";

import {
  SellerSegmentGuard,
  SellerStatusShell,
} from "@/modules/seller";

export default function SellerSuspendedPage() {
  return (
    <SellerSegmentGuard allow={["suspended"]}>
      <SellerStatusShell kind="suspended" />
    </SellerSegmentGuard>
  );
}
