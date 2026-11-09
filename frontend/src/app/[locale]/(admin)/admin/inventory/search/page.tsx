"use client";

import { Suspense } from "react";
import { AdminInventorySearch } from "@/modules/admin/routes/inventorySearch";

export default function AdminInventorySearchPage() {
	return (
		<Suspense>
			<AdminInventorySearch />
		</Suspense>
	);
}
