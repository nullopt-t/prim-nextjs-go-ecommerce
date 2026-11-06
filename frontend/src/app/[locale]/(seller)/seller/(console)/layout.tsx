"use client";

import { SellerConsoleGate } from "@/modules/seller";

export default function SellerConsoleLayout({ children }: { children: React.ReactNode }) {
  return <SellerConsoleGate>{children}</SellerConsoleGate>;
}
