import RecentlyCard from "@/features/home/components/ui/recentlyCard";

export default function RecentlyGrid() {
	return (
		<div className="flex flex-wrap gap-5 mt-5">
			<RecentlyCard title="smart watches" />
			<RecentlyCard title="smart watches" />
			<RecentlyCard title="smart watches" />
			<RecentlyCard title="smart watches" />
			<RecentlyCard title="smart watches" />
		</div>
	);
}
