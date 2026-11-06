"use client";

import {
  VendorSegmentGuard,
  VendorStatusShell,
} from "@/modules/vendor";

export default function VendorSuspendedPage() {
  return (
    <VendorSegmentGuard allow={["suspended"]}>
      <VendorStatusShell kind="suspended" />
    </VendorSegmentGuard>
  );
}
