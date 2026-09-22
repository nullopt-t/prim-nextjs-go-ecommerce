import { Brand } from "@/components/ui/Brand";
import { SearchBar } from "@/components/ui/search";
import { HeaderActions } from "@/components/ui/headerActions";

export function Header() {
  return (
    <header className="sticky top-0 z-50 text-foreground bg-background/80 backdrop-blur-xl border-b border-border/60 transition-colors">
      <div className="px-4 py-2.5 md:px-6 md:py-3 lg:px-10 lg:py-4 flex flex-col md:grid md:grid-cols-3 md:items-center gap-2.5 md:gap-6">
        {/* Top row on mobile: Brand on left, Actions on right */}
        <div className="flex items-center justify-between w-full md:w-auto">
          <Brand />
          <div className="flex md:hidden items-center">
            <HeaderActions />
          </div>
        </div>

        {/* Search bar: full-width second row on mobile, centered on desktop */}
        <div className="w-full">
          <SearchBar />
        </div>

        {/* Actions: visible on desktop, placed on the right */}
        <div className="hidden md:flex justify-end items-center gap-5 justify-self-end text-foreground">
          <HeaderActions />
        </div>
      </div>
    </header>
  );
}
