"use client";

import { VendorLayout } from "@/modules/vendor";
import { RequireRole } from "@/shared/ui/requireRole";

export default function VendorRootLayout({ children }: { children: React.ReactNode }) {
	// Any authenticated shopper may enter the vendor shell; VendorContext
	// decides between the dashboard (profile exists) and the registration
	// form (no profile yet). Requiring the vendor role here would lock out
	// pending merchants whose user role is still customer.
	return (
		<RequireRole minimumRole="customer">
			<VendorLayout>{children}</VendorLayout>
		</RequireRole>
	);
}
