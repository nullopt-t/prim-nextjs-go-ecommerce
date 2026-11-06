export default function ShopLoading() {
	return (
		<div role="status" aria-live="polite" aria-busy="true" className="w-full space-y-4 py-4">
			<div className="h-8 w-56 animate-pulse rounded-md bg-secondary/60" />
			<div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
				{Array.from({ length: 8 }).map((_, index) => (
					<div key={index} className="space-y-2">
						<div className="aspect-square animate-pulse rounded-lg bg-secondary/50" />
						<div className="h-3 w-3/4 animate-pulse rounded bg-secondary/60" />
						<div className="h-3 w-1/3 animate-pulse rounded bg-secondary/60" />
					</div>
				))}
			</div>
		</div>
	);
}
