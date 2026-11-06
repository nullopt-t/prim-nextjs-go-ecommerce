"use client";

import { SellerProvider } from "@/modules/seller";
import { RequireAuth } from "@/shared/ui/requireAuth";

export default function SellerRootLayout({ children }: { children: React.ReactNode }) {
  return (
    <RequireAuth>
      <SellerProvider>{children}</SellerProvider>
    </RequireAuth>
  );
}
