import { Brand } from "@/components/ui/Brand";
import { SearchBar } from "@/components/ui/search";
import { HeaderActions } from "@/components/ui/headerActions";
import { MobileSearchDialog } from "@/components/ui/mobileSearchDialog";

export function Header() {
	return (
		<header className="sticky top-0 z-50 w-full bg-background/85 backdrop-blur-md border-b border-border/80 shadow-xs transition-colors">
			<div className="w-full px-3 sm:px-6 lg:px-8 h-15 sm:h-16 flex items-center justify-between gap-3 sm:gap-6">
				<div className="flex-shrink-0">
					<Brand />
				</div>
				<div className="flex-1 max-w-2xl hidden md:block">
					<SearchBar />
				</div>
				<div className="flex items-center gap-2 flex-shrink-0">
					<MobileSearchDialog />
					<HeaderActions />
				</div>
			</div>
		</header>
	);
}
