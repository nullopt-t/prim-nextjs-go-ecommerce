"use client";

import { AdminLayout } from "@/modules/admin";
import { RequireRole } from "@/shared/ui/requireRole";

export default function AdminRootLayout({ children }: { children: React.ReactNode }) {
	return (
		<RequireRole minimumRole="admin">
			<AdminLayout>{children}</AdminLayout>
		</RequireRole>
	);
}
