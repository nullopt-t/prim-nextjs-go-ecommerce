"use client";

import { AlertTriangle, RefreshCw } from "lucide-react";
import { useTranslations } from "next-intl";

/**
 * Standard in-view error state for content that failed to load.
 *
 * Use this instead of a toast for mount-time / data-loading failures so the
 * user sees why the surface is empty and can retry, rather than a notification
 * that disappears leaving a blank view behind.
 */
export function ErrorState({
  onRetry,
  title,
  description,
  className = "",
}: {
  onRetry?: () => void;
  title?: string;
  description?: string;
  className?: string;
}) {
  const t = useTranslations("common.errors");

  return (
    <div
      className={`flex flex-col items-center justify-center gap-4 rounded-2xl border border-border bg-background p-10 text-center shadow-xs ${className}`}
    >
      <div className="rounded-full bg-destructive/10 p-3 text-destructive">
        <AlertTriangle className="size-6" />
      </div>
      <div className="space-y-1">
        <p className="text-sm font-bold text-foreground">{title || t("loadTitle")}</p>
        <p className="mx-auto max-w-sm text-xs text-muted-foreground">
          {description || t("loadDescription")}
        </p>
      </div>
      {onRetry && (
        <button
          type="button"
          onClick={onRetry}
          className="mt-1 inline-flex items-center gap-2 rounded-xl bg-accent-brand px-4 py-2 text-xs font-semibold text-white transition-colors hover:bg-accent-brand/90"
        >
          <RefreshCw className="size-4" />
          {t("retry")}
        </button>
      )}
    </div>
  );
}
