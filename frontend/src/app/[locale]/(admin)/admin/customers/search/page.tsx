"use client";

import { Suspense } from "react";
import { AdminCustomersSearch } from "@/modules/admin/routes/customersSearch";

export default function AdminCustomersSearchPage() {
	return (
		<Suspense>
			<AdminCustomersSearch />
		</Suspense>
	);
}
