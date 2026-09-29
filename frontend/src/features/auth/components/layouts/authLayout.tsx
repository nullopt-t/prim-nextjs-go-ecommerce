"use client";

import { Link } from "@/i18n/navigation";
import { Sparkles, House, Moon, Sun } from "lucide-react";
import { useTheme } from "@/context/theme";

export function AuthLayout({ children }: { children: React.ReactNode }) {
  const { theme, toggle } = useTheme();

  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col justify-between p-4 sm:p-6 md:p-8 selection:bg-accent-brand/20">
      {/* Top clean navigation bar */}
      <header className="w-full max-w-5xl mx-auto flex items-center justify-between py-2">
        <Link
          href="/"
          className="flex items-center gap-2 hover:opacity-85 transition-opacity"
        >
          <div className="size-8 rounded-md bg-accent-brand text-white flex items-center justify-center shadow-xs">
            <Sparkles className="size-4" />
          </div>
          <span className="font-bold text-lg tracking-tight text-foreground">
            PRIM
          </span>
        </Link>

        <div className="flex items-center gap-2">
          <Link
            href="/"
            aria-label="Home"
            className="p-2 text-muted-foreground hover:text-foreground hover:bg-secondary rounded-md transition-colors inline-flex items-center justify-center"
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
      </header>

      {/* Main centered card */}
      <main className="w-full flex-1 flex items-center justify-center py-8">
        <div className="w-full max-w-[420px] bg-card border border-border rounded-md p-6 sm:p-8 shadow-xs">
          {children}
        </div>
      </main>

      {/* Bottom subtle footer */}
      <footer className="w-full max-w-5xl mx-auto text-center py-2 text-xs text-muted-foreground">
        © {new Date().getFullYear()} PRIM. All rights reserved.
      </footer>
    </div>
  );
}

export default AuthLayout;
