"use client";

import { useEffect, useState } from "react";
import { AlertCircle, X } from "lucide-react";
import { usePathname } from "next/navigation";

/**
 * Embedded in-page action banner.
 *
 * Replaces floating toasts with an in-page, embedded error banner.
 * Success actions are silent; errors are visibly embedded in the page flow.
 */
export function ResponsiveToaster({
  className = "",
}: {
  richColors?: boolean;
  className?: string;
}) {
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const pathname = usePathname();

  // Clear embedded error on route navigation
  useEffect(() => {
    setErrorMessage(null);
  }, [pathname]);

  useEffect(() => {
    const handleEmbeddedError = (event: Event) => {
      const customEvent = event as CustomEvent<{ message: string | null }>;
      setErrorMessage(customEvent.detail?.message || null);
    };

    window.addEventListener("prim:embedded-error", handleEmbeddedError);
    return () => window.removeEventListener("prim:embedded-error", handleEmbeddedError);
  }, []);

  if (!errorMessage) return null;

  return (
    <div className={`fixed top-4 left-1/2 -translate-x-1/2 z-50 w-full max-w-xl px-4 pointer-events-none ${className}`}>
      <div
        role="alert"
        aria-live="assertive"
        className="flex items-center justify-between gap-3 rounded-2xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm font-semibold text-destructive shadow-sm backdrop-blur-md transition-all pointer-events-auto"
      >
        <div className="flex items-center gap-2.5">
          <AlertCircle className="size-5 shrink-0" aria-hidden="true" />
          <span>{errorMessage}</span>
        </div>
        <button
          type="button"
          onClick={() => setErrorMessage(null)}
          aria-label="Dismiss error"
          className="rounded-lg p-1 text-destructive/80 transition-colors hover:bg-destructive/15 hover:text-destructive cursor-pointer"
        >
          <X className="size-4" aria-hidden="true" />
        </button>
      </div>
    </div>
  );
}
