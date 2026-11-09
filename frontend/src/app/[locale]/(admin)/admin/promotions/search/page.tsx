"use client";

import { Suspense } from "react";
import { AdminPromotionsSearch } from "@/modules/admin/routes/promotionsSearch";

export default function AdminPromotionsSearchPage() {
	return (
		<Suspense>
			<AdminPromotionsSearch />
		</Suspense>
	);
}
