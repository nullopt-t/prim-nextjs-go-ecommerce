"use client";

import { Brand } from "@/components/ui/Brand";
import { Link } from "@/i18n/navigation";
import { House, Moon, Sun } from "lucide-react";
import { useTheme } from "@/context/theme";

export function AuthLayout({ children }: { children: React.ReactNode }) {
  const { theme, toggle } = useTheme();

  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col justify-between selection:bg-accent-brand/20">
      {/* Top header taking the exact same properties as Main Header (h-15 sm:h-16, px-3 sm:px-6 lg:px-8, no margins) */}
      <header className="sticky top-0 z-50 w-full bg-background/85 backdrop-blur-md border-b border-border/80 shadow-xs transition-colors">
        <div className="w-full px-3 sm:px-6 lg:px-8 h-15 sm:h-16 flex items-center justify-between">
          <Brand />

          <div className="flex items-center gap-2">
            <Link
              href="/"
              aria-label="Home"
              className="p-2 text-muted-foreground hover:text-foreground hover:bg-secondary rounded-md transition-colors inline-flex items-center justify-center cursor-pointer"
            >
              <House className="size-4" />
            </Link>
            <button
              type="button"
              onClick={toggle}
              aria-label="Toggle theme"
              className="p-2 text-muted-foreground hover:text-foreground hover:bg-secondary rounded-md transition-colors cursor-pointer inline-flex items-center justify-center"
            >
              {theme === "light" ? (
                <Sun className="size-4" />
              ) : (
                <Moon className="size-4" />
              )}
            </button>
          </div>
        </div>
      </header>

      {/* Main flat content taking exact main layout padding properties (px-3 sm:px-6 lg:px-8 py-4 sm:py-6) */}
      <main className="w-full px-3 sm:px-6 lg:px-8 py-6 sm:py-10 flex-1 flex flex-col justify-center">
        {children}
      </main>

      {/* Bottom footer matching exact header horizontal padding */}
      <footer className="w-full px-3 sm:px-6 lg:px-8 text-center py-4 text-xs text-muted-foreground border-t border-border/50">
        © {new Date().getFullYear()} PRIM. All rights reserved.
      </footer>
    </div>
  );
}

export default AuthLayout;
