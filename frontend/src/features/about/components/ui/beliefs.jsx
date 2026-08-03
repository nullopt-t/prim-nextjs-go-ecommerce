import BeliefsGrid from "@/features/about/components/ui/beliefsGrid";
import BeliefsDescription from "@/features/about/components/ui/beliefsDescription";

export default function Beliefs() {
	return (
		<div className="flex flex-col lg:flex-row gap-10 justify-between">
			<div className="lg:w-96">
				<BeliefsDescription />
			</div>
			<div className="">
				<BeliefsGrid />
			</div>
		</div>
	);
}
