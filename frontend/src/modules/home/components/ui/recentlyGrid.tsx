import RecentlyCard from "@/modules/home/components/ui/recentlyCard";
import type { ViewedProduct } from "@/modules/history/types";

export default function RecentlyGrid({ items }: { items: ViewedProduct[] }) {
	return (
		<div className="flex gap-4 sm:gap-6 mt-4 overflow-x-auto pb-2 scrollbar-none snap-x">
			{items.map((item) => (
				<div key={item.productId} className="shrink-0">
					<RecentlyCard
						title={item.title}
						img={item.thumbnailUrl}
						slug={item.slug}
					/>
				</div>
			))}
		</div>
	);
}
