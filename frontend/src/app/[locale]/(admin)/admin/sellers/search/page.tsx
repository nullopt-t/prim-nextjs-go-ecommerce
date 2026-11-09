"use client";

import { Suspense } from "react";
import { AdminSellersSearch } from "@/modules/admin/routes/sellersSearch";

export default function AdminSellersSearchPage() {
	return (
		<Suspense>
			<AdminSellersSearch />
		</Suspense>
	);
}
