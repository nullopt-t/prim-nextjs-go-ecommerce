"use client";

import { ChevronLeft, ChevronRight } from "lucide-react";
import { useTranslations } from "next-intl";
import { useCallback, useState } from "react";

import { CursorMeta, emptyCursorMeta } from "@/lib/pagination";

export type { CursorMeta };
export { emptyCursorMeta };

/**
 * Client-side navigation state for forward-only keyset pagination.
 *
 * The API only hands back a cursor for the *next* page, so we keep the chain of
 * cursors walked so far. "Previous" simply pops the chain and re-queries, which
 * avoids needing a second (reverse) query on the backend.
 */
export function useCursorPagination() {
  const [stack, setStack] = useState<string[]>([""]);
  const cursor = stack[stack.length - 1];
  const hasPrev = stack.length > 1;

  const next = useCallback((nextCursor: string) => {
    if (!nextCursor) return;
    setStack((s) => [...s, nextCursor]);
  }, []);

  const prev = useCallback(() => {
    setStack((s) => (s.length > 1 ? s.slice(0, -1) : s));
  }, []);

  const reset = useCallback(() => setStack([""]), []);

  return { cursor, hasPrev, next, prev, reset, page: stack.length };
}

interface CursorPagerProps {
  hasPrev: boolean;
  hasMore: boolean;
  page: number;
  onPrev: () => void;
  onNext: () => void;
  loading?: boolean;
  /** Number of rows currently rendered, shown on the start-aligned label. */
  shown?: number;
}

/** Prev/next footer for a keyset-paginated table. */
export function CursorPager({
  hasPrev,
  hasMore,
  page,
  onPrev,
  onNext,
  loading,
  shown,
}: CursorPagerProps) {
  const t = useTranslations("admin.pagination");

  return (
    <div className="p-4 border-t border-border/80 bg-muted/10 flex items-center justify-between gap-4 text-xs text-muted-foreground">
      <span>{typeof shown === "number" ? t("showing", { shown }) : null}</span>

      <div className="flex items-center gap-2">
        <button
          onClick={onPrev}
          disabled={!hasPrev || loading}
          className="p-1.5 rounded-lg border border-border bg-card hover:bg-secondary text-foreground disabled:opacity-40 transition-colors"
          title={t("prev")}
          aria-label={t("prev")}
        >
          <ChevronLeft className="size-4 rtl:rotate-180" />
        </button>
        <span className="font-mono text-xs px-2 font-bold text-foreground">
          {t("page", { page })}
        </span>
        <button
          onClick={onNext}
          disabled={!hasMore || loading}
          className="p-1.5 rounded-lg border border-border bg-card hover:bg-secondary text-foreground disabled:opacity-40 transition-colors"
          title={t("next")}
          aria-label={t("next")}
        >
          <ChevronRight className="size-4 rtl:rotate-180" />
        </button>
      </div>
    </div>
  );
}
