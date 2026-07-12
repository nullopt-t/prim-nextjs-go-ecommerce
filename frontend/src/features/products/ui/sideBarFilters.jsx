import { SlidersHorizontal } from "lucide-react";

function SideBarFilters() {
	return (
		<div className="flex justify-between text-txt-sm md:text-txt-md lg:text-txt-lg text-muted-foreground">
			<div className="flex gap-1">
				<SlidersHorizontal />
				<p className="">Filters</p>
			</div>
			<p className="cursor-pointer">Clear all</p>
		</div>
	);
}

export default SideBarFilters;
