import { Brand, Categories, SearchBar, HeaderActions } from "@/components/ui";

export function Header() {
	return (
		<div className="p-2.5 md:p-5 lg:p-10 sticky top-0 z-50 text-foreground bg-background/70 backdrop-blur-3xl">
			<div className="pb-2.5 grid gap-5 grid-cols-3 border-b border-border">
				<Brand />
				<div className="col-span-3 md:col-span-1">
					<SearchBar />
				</div>
				<div className="col-span-3 md:col-span-1 justify-self-end flex justify-end items-center gap-5 text-foreground">
					<HeaderActions />
				</div>
			</div>
			<div className="pt-2.5">
				<Categories />
			</div>
		</div>
	);
}
