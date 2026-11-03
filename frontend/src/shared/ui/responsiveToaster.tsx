"use client";

import { Toaster } from "sonner";
import { useEffect, useState } from "react";
import { useTheme } from "@/shared/theme";

/**
 * Renders Sonner toasts styled to match PRIM theme (light & dark mode,
 * colors, borders, typography, and button styles).
 *
 * - Phones (< 768px):  top-right  (keeps bottom nav reachable)
 * - Tablet+ (≥ 768px): bottom-right (desktop convention)
 */
export function ResponsiveToaster({
  richColors = true,
  className = "",
}: {
  richColors?: boolean;
  className?: string;
}) {
  const { theme } = useTheme();
  const [position, setPosition] = useState<"top-right" | "bottom-right">(
    "bottom-right"
  );

  useEffect(() => {
    const update = () => {
      setPosition(window.innerWidth < 768 ? "top-right" : "bottom-right");
    };
    update();
    window.addEventListener("resize", update);
    return () => window.removeEventListener("resize", update);
  }, []);

  return (
    <Toaster
      position={position}
      theme={theme as "light" | "dark"}
      richColors={richColors}
      className={className}
      closeButton={true}
      toastOptions={{
        classNames: {
          toast:
            "group font-sans border shadow-lg rounded-xl text-xs sm:text-sm transition-all duration-200",
          title: "font-semibold text-foreground text-xs sm:text-sm",
          description: "text-muted-foreground text-[11px] sm:text-xs",
          actionButton:
            "cursor-pointer font-medium text-xs rounded-lg px-3 py-1.5 transition-colors !bg-accent-brand !text-white hover:opacity-90 active:scale-95 shadow-2xs",
          cancelButton:
            "cursor-pointer font-medium text-xs rounded-lg px-3 py-1.5 transition-colors !bg-secondary !text-secondary-foreground hover:bg-secondary/80",
          closeButton:
            "cursor-pointer border border-border/80 bg-background text-muted-foreground hover:text-foreground transition-colors",
        },
      }}
    />
  );
}
