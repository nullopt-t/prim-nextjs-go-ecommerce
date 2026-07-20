import { Categories } from "@/components/ui/categories";
import { SearchBar } from "@/components/ui/search";
import { HeaderActions } from "@/components/ui/headerActions";

export function Header() {
	return (
		<div className="sticky top-0 z-50 text-foreground bg-background/70 backdrop-blur-3xl">
			<div className="grid gap-5 grid-cols-3 p-5 pr-2.5 pl-2.5 border-b border-border-color">
				<h1 className="col-span-3 md:col-span-1 md:justify-self-start font-medium text-title-sm md:text-title-md lg:text-title-lg text-center">
					<span className="">PRI</span>
					<span className="text-orange-500">M</span>
				</h1>
				<div className="col-span-3 md:col-span-1">
					<SearchBar />
				</div>
				<div className="col-span-3 md:col-span-1 justify-self-end flex justify-end items-center gap-5 text-foreground">
					<HeaderActions />
				</div>
			</div>
			<div className="p-5 pr-2.5 pl-2.5">
				<Categories />
			</div>
		</div>
	);
}
