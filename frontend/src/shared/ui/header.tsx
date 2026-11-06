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
			<div className="w-full px-3 sm:px-6 lg:px-8 h-15 sm:h-16 flex items-center justify-between gap-3 sm:gap-6">
				<div className="flex items-center gap-2">
					<MobileNav />
					<div className="flex-shrink-0">
						<Brand />
					</div>
				</div>
				{centerContent !== undefined ? (
					<div className="flex-1 hidden md:flex items-center justify-start min-w-0">
						{centerContent}
					</div>
				) : (
					<div className="flex-1 max-w-2xl hidden md:block">
						<SearchBar />
					</div>
				)}
				<div className="flex items-center gap-2 flex-shrink-0">
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
