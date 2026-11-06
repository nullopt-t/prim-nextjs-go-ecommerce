"use client";

import { VendorRegister, VendorSegmentGuard } from "@/modules/vendor";

export default function VendorRegisterPage() {
  return (
    <VendorSegmentGuard allow={["unregistered", "rejected"]}>
      <VendorRegister />
    </VendorSegmentGuard>
  );
}
