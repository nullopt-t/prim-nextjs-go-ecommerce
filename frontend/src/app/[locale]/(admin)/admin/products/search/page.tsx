"use client";

import { Suspense } from "react";
import { AdminProductsSearch } from "@/modules/admin/routes/productsSearch";

export default function AdminProductsSearchPage() {
	return (
		<Suspense>
			<AdminProductsSearch />
		</Suspense>
	);
}
