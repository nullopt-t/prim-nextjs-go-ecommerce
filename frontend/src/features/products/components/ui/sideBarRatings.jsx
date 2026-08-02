import Stars from "@/components/ui/stars";
import SideBarTitle from "@/features/products/components/ui/sideBarTitle";

function SideBarRatings() {
	const rates = [
		{
			id: "stars-2349u0",
			stars: "5",
		},
		{
			id: "stars-sfasfv",
			stars: "4",
		},
		{
			id: "stars-afsafsavs",
			stars: "3",
		},
		{
			id: "stars-sfsafsv",
			stars: "2",
		},
		{
			id: "stars-34325t",
			stars: "1",
		},
	];

	return (
		<div className="border-b border-border pb-5">
			<SideBarTitle title="ratings" />
			<div className="flex flex-col gap-2.5">
				{rates.map((value) => (
					<div
						key={value.id}
						className="flex gap-2.5 items-center group cursor-pointer"
					>
						<Stars starsNum={value.stars} />
						<p className="text-muted-foreground group-hover:text-accent-brand">
							<span className="">&amp;</span>{" "}
							<span className="">up</span>
						</p>
					</div>
				))}
			</div>
		</div>
	);
}

export default SideBarRatings;
