"use client";

import { VendorLayout } from "@/modules/vendor";
import { RequireRole } from "@/shared/ui/requireRole";

export default function VendorRootLayout({ children }: { children: React.ReactNode }) {
	return (
		<RequireRole minimumRole="vendor">
			<VendorLayout>{children}</VendorLayout>
		</RequireRole>
	);
}
