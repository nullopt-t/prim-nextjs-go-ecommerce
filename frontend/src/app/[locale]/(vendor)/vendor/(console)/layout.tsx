"use client";

import { VendorConsoleGate } from "@/modules/vendor";

export default function VendorConsoleLayout({ children }: { children: React.ReactNode }) {
  return <VendorConsoleGate>{children}</VendorConsoleGate>;
}
