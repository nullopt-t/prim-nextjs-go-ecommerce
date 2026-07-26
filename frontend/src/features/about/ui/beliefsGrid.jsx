import { Star, Gauge, ShieldCheck, Smile } from "lucide-react";
import BeliefCards from "@/features/about/ui/beliefsCards";

export default function BeliefsGrid() {
	const beliefsArr = [
		{
			id: "quality-cdkt",
			title: "Quality",
			subTitle:
				"Every product is vetted for excellence before it reaches you.",
			icon: Star,
		},
		{
			id: "speed-cdkt",
			title: "Speed",
			subTitle:
				"Same-day dispatch and next-day delivery across the region.",
			icon: Gauge,
		},
		{
			id: "trust-cdkt",
			title: "Trust",
			subTitle:
				"Transparent pricing, honest reviews, and no hidden fees.",
			icon: ShieldCheck,
		},
		{
			id: "Simplicity-cdkt",
			title: "Simplicity",
			subTitle:
				"Shopping should feel effortless — we obsess over the details",
			icon: Smile,
		},
	];

	return (
		<div className="grid grid-cols-2 gap-5">
			{beliefsArr.map((belief) => (
				<BeliefCards key={belief.id} belief={belief} />
			))}
		</div>
	);
}
