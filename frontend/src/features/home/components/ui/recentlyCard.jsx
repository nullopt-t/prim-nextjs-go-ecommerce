export default function RecentlyCard({ title }) {
	return (
		<div className="w-fit cursor-pointer group">
			<div className="w-24 h-24 bg-border rounded-md hover:bg-border/45"></div>
			<p className="group-hover:text-accent-brand text-[12px] text-muted-foreground mt-2.5 text-center capitalize">
				{title}
			</p>
		</div>
	);
}
