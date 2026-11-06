"use client";

import { SellerRegister, SellerSegmentGuard } from "@/modules/seller";

export default function SellerRegisterPage() {
  return (
    <SellerSegmentGuard allow={["unregistered", "rejected"]}>
      <SellerRegister />
    </SellerSegmentGuard>
  );
}
