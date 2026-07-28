import RecentlyGrid from "@/features/home/components/ui/recentlyGrid";

export default function Recently() {
	return (
		<div className="bg-border/25 p-5 m-2.5 mb-0 md:mb-0 lg:mb-0 md:m-5 lg:m-10 border-t border-border">
			<p className="text-muted-foreground font-medium text-sm">
				Recently Viewed
			</p>
			<RecentlyGrid />
		</div>
	);
}
