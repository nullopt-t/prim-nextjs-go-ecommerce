import { Brand } from "@/components/ui/Brand";
import { SearchBar } from "@/components/ui/search";
import { HeaderActions } from "@/components/ui/headerActions";

export function Header() {
  return (
    <header className="p-2.5 md:p-5 lg:px-10 lg:py-4 sticky top-0 z-50 text-foreground bg-background/70 backdrop-blur-3xl border-b border-border/50">
      <div className="grid gap-5 grid-cols-3 items-center">
        <Brand />
        <div className="col-span-3 md:col-span-1">
          <SearchBar />
        </div>
        <div className="col-span-3 md:col-span-1 justify-self-end flex justify-end items-center gap-5 text-foreground">
          <HeaderActions />
        </div>
      </div>
    </header>
  );
}
