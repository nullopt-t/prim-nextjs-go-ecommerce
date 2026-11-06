"use client";

import { VendorProvider } from "@/modules/vendor";
import { RequireAuth } from "@/shared/ui/requireAuth";

export default function VendorRootLayout({ children }: { children: React.ReactNode }) {
  return (
    <RequireAuth>
      <VendorProvider>{children}</VendorProvider>
    </RequireAuth>
  );
}
