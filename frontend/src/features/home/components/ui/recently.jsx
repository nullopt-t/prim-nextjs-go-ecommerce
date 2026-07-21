import RecentlyGrid from "@/features/home/components/ui/recentlyGrid";

export default function Recently() {
	return (
		<div className="bg-border/25 p-5 border-t border-border">
			<p className="text-muted-foreground font-medium text-sm">
				Recently Viewed
			</p>
			<RecentlyGrid />
		</div>
	);
}
