"use client";

import { ErrorState } from "@/shared/ui/errorState";

export default function ShopError({ reset }: { error: Error; reset: () => void }) {
	return (
		<div className="py-8">
			<ErrorState onRetry={reset} />
		</div>
	);
}
