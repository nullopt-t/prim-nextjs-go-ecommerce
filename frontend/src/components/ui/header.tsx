import { Brand } from "@/components/ui/Brand";
import { SearchBar } from "@/components/ui/search";
import { HeaderActions } from "@/components/ui/headerActions";

export function Header() {
	return (
		<header className="sticky top-0 z-50 w-full bg-background/85 backdrop-blur-md border-b border-border/80 shadow-xs transition-colors">
			<div className="max-w-screen-2xl mx-auto px-4 sm:px-6 lg:px-10 h-15 sm:h-16 flex items-center justify-between gap-3 sm:gap-6">
				<div className="flex-shrink-0">
					<Brand />
				</div>
				<div className="flex-1 max-w-2xl hidden md:block">
					<SearchBar />
				</div>
				<div className="flex items-center flex-shrink-0">
					<HeaderActions />
				</div>
			</div>
			{/* Mobile Search - shown only on small screens */}
			<div className="md:hidden px-4 sm:px-6 pb-2.5 pt-0.5 max-w-screen-2xl mx-auto w-full">
				<SearchBar />
			</div>
		</header>
	);
}
