"use client";

import { useLocale, useTranslations } from "next-intl";
import { formatDateTime } from "@/shared/lib/datetime";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import type { OrderStatusHistoryEntry } from "@/modules/orders/types";

interface OrderTimelineHistoryProps {
  history?: OrderStatusHistoryEntry[];
  createdAt?: string;
  className?: string;
  title?: string;
}

export function OrderTimelineHistory({
  history,
  createdAt,
  className = "",
  title,
}: OrderTimelineHistoryProps) {
  const t = useTranslations("orderDetails");
  const tOrders = useTranslations("orders");
  const locale = useLocale();

  const rawEntries = Array.isArray(history) ? history : [];

  // If no history records exist, synthesize at least the initial placement event
  const entries: Array<{
    id: string;
    status: string;
    notes?: string | null;
    createdAt: string;
  }> =
    rawEntries.length > 0
      ? rawEntries
      : createdAt
      ? [
          {
            id: "initial-order-placed",
            status: "pending",
            notes: null,
            createdAt,
          },
        ]
      : [];

  // Sort ascending (chronological: oldest to newest) for timeline flow
  const sorted = [...entries].sort((a, b) => {
    const tA = new Date(a.createdAt).getTime();
    const tB = new Date(b.createdAt).getTime();
    return tA - tB;
  });

  if (sorted.length === 0) return null;

  return (
    <section className={`rounded-2xl border border-border bg-card p-5 shadow-xs sm:p-6 ${className}`}>
      <h3 className="text-base font-semibold text-foreground mb-4">
        {title || t("timeline.title", { defaultMessage: "Timeline & Status History" })}
      </h3>

      <div className="relative ps-6 space-y-6 before:absolute before:start-2.5 before:top-2 before:bottom-2 before:w-0.5 before:bg-border/70">
        {sorted.map((item, idx) => {
          const isLatest = idx === sorted.length - 1;
          const statusKey = String(item.status || "pending").toLowerCase();
          const label = tOrders(`status.${statusKey}`, {
            defaultMessage: statusKey.charAt(0).toUpperCase() + statusKey.slice(1),
          });
          const formattedDate = formatDateTime(item.createdAt, locale);

          return (
            <div key={item.id || idx} className="relative flex flex-col gap-1.5 group">
              {/* Timeline Node Bullet */}
              <div
                className={`absolute -start-6 top-1.5 size-3 rounded-full border-2 transition-transform ${
                  isLatest
                    ? "border-accent-brand bg-accent-brand ring-4 ring-accent-brand/15 scale-110"
                    : "border-border bg-background"
                }`}
              />

              <div className="flex flex-wrap items-center gap-2">
                <StatusPill tone={statusTone(statusKey)}>{label}</StatusPill>
                <time className="text-xs text-muted-foreground">{formattedDate}</time>
              </div>

              {item.notes ? (
                <div className="mt-1 rounded-xl bg-secondary/50 border border-border/60 p-3 text-xs sm:text-sm text-foreground leading-relaxed whitespace-pre-wrap">
                  {item.notes}
                </div>
              ) : null}
            </div>
          );
        })}
      </div>
    </section>
  );
}
