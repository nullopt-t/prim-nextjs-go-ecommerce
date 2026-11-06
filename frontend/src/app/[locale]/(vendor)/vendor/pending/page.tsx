"use client";

import {
  VendorSegmentGuard,
  VendorStatusShell,
} from "@/modules/vendor";

export default function VendorPendingPage() {
  return (
    <VendorSegmentGuard allow={["pending"]}>
      <VendorStatusShell kind="pending" />
    </VendorSegmentGuard>
  );
}
