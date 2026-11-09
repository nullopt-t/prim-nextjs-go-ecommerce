import { Brand } from "@/shared/ui/Brand";
import { DesktopNav, MobileNav } from "@/shared/ui/mainNav";
import { SearchBar } from "@/shared/ui/search";
import { HeaderActions } from "@/shared/ui/headerActions";
import { MobileSearchDialog } from "@/shared/ui/mobileSearchDialog";

interface HeaderProps {
	centerContent?: React.ReactNode;
	actionsContext?: "default" | "settings";
}

export function Header({ centerContent, actionsContext = "default" }: HeaderProps = {}) {
	return (
		<header className="sticky top-0 z-50 w-full bg-background/85 backdrop-blur-md border-b border-border/80 shadow-xs transition-colors">
		<div className="w-full px-2 sm:px-6 lg:px-8 flex items-center justify-between gap-1 sm:gap-4 h-14 sm:h-16">
			{/* Left: Mobile Nav & Brand Logo */}
			<div className="flex items-center gap-1 sm:gap-2 shrink-0">
				<MobileNav />
				<Brand />
			</div>

			{/* Center: Search Bar on md+ */}
			{centerContent !== undefined ? (
				<div className="flex-1 hidden md:flex items-center justify-start min-w-0 px-4">
					{centerContent}
				</div>
			) : (
				<div className="flex-1 max-w-2xl hidden md:block px-4">
					<SearchBar />
				</div>
			)}

			{/* Right: Actions, Search Trigger, Cart, Currency, Account */}
			<div className="flex items-center justify-end gap-1 sm:gap-2 shrink-0">
				<MobileSearchDialog />
				<HeaderActions context={actionsContext} />
			</div>
		</div>
		<div className="hidden md:block border-t border-border/60">
			<div className="w-full px-3 sm:px-6 lg:px-8 py-2">
				<DesktopNav />
			</div>
		</div>
		</header>
	);
}
