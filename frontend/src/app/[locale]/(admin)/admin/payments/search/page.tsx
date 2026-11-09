"use client";

import { Suspense } from "react";
import { AdminPaymentsSearch } from "@/modules/admin/routes/paymentsSearch";

export default function AdminPaymentsSearchPage() {
	return (
		<Suspense>
			<AdminPaymentsSearch />
		</Suspense>
	);
}
