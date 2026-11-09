"use client";

import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { useState, useEffect, useCallback, useRef } from "react";
import { useSearchParams } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { adminService } from "@/modules/admin/api/adminService";
import type { Order, OrderItem } from "@/modules/orders/types";
import { formatDateTime } from "@/shared/lib/datetime";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { CursorPager, useCursorPagination } from "@/shared/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/shared/lib/pagination";
import { Link } from "@/shared/i18n/navigation";

interface OrderView {
  id: string;
  customer: string;
  email: string;
  date: string;
  total: string;
  discount: string | null;
  couponCode: string | null;
  status: string;
  fulfillmentStatus: string;
  paymentStatus: string;
  items: OrderItem[];
}

const KNOWN_STATUSES = ["all", "pending", "processing", "shipped", "delivered", "cancelled"];

function toISODate(d: Date): string {
  return d.toISOString().slice(0, 10);
}

function applyPreset(preset: string): { from: string; to: string } {
  const now = new Date();
  const today = toISODate(now);
  switch (preset) {
    case "today":
      return { from: today, to: today };
    case "7d": {
      const d = new Date(now);
      d.setDate(d.getDate() - 6);
      return { from: toISODate(d), to: today };
    }
    case "30d": {
      const d = new Date(now);
      d.setDate(d.getDate() - 29);
      return { from: toISODate(d), to: today };
    }
    default:
      return { from: "", to: "" };
  }
}

export function AdminOrdersSearch() {
  const searchParams = useSearchParams();
  const initialStatus = (() => {
    const s = (searchParams.get("status") || "all").toLowerCase();
    return KNOWN_STATUSES.includes(s) ? s : "all";
  })();
  const t = useTranslations("admin");
  const locale = useLocale();
  const [orders, setOrders] = useState<OrderView[]>([]);
  const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);

  // Search & Filter parameters (backend-driven)
  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState(initialStatus);
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");
  const [datePreset, setDatePreset] = useState("all");

  // Status popup state
  const [statusOpen, setStatusOpen] = useState(false);
  const statusRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const id = setTimeout(() => setDebouncedSearch(searchQuery.trim()), 300);
    return () => clearTimeout(id);
  }, [searchQuery]);

  const handleSearchKey = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      setDebouncedSearch(searchQuery.trim());
      reset();
    }
  };

  useEffect(() => {
    if (!statusOpen) return;
    const onDown = (e: MouseEvent) => {
      if (statusRef.current && !statusRef.current.contains(e.target as Node)) setStatusOpen(false);
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setStatusOpen(false);
    };
    document.addEventListener("mousedown", onDown);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onDown);
      document.removeEventListener("keydown", onKey);
    };
  }, [statusOpen]);

  const statusOptions = [
    { value: "all", label: t("status.all") },
    { value: "pending", label: t("status.pending") },
    { value: "processing", label: t("status.processing") },
    { value: "shipped", label: t("status.shipped") },
    { value: "delivered", label: t("status.delivered") },
    { value: "cancelled", label: t("status.cancelled") },
  ];
  const statusLabel = statusOptions.find((o) => o.value === statusFilter)?.label ?? statusFilter;

  const { cursor, hasPrev, next, prev, reset, page } = useCursorPagination();

  const loadOrders = useCallback(async () => {
    try {
      setLoading(true);
      setLoadError(false);
      const res = await adminService.getOrders({
        cursor: cursor || undefined,
        pageSize: 50,
        search: debouncedSearch || undefined,
        status: statusFilter !== "all" ? statusFilter.toLowerCase() : undefined,
        dateFrom: dateFrom || undefined,
        dateTo: dateTo || undefined,
      });
      const list = res?.data ?? [];
      const formatted = list.map((so: Order) => ({
        id: so.id,
        customer: so.customerEmail ? so.customerEmail.split("@")[0] : "—",
        email: so.customerEmail || "—",
        date: formatDateTime(so.createdAt, locale),
        total: `EGP ${((so.totalAmount || 0) / 100).toFixed(2)}`,
        discount: so.discountAmount ? `EGP ${((so.discountAmount || 0) / 100).toFixed(2)}` : null,
        couponCode: so.couponCode || null,
        status: so.status ? so.status.toLowerCase() : "pending",
        fulfillmentStatus: so.status === "shipped" || so.status === "delivered" ? "Fulfilled" : "Unfulfilled",
        paymentStatus: so.paymentStatus ? String(so.paymentStatus).toLowerCase() : "unknown",
        items: so.items || [],
      }));
      setOrders(formatted);
      if (res?.meta) setMeta(res.meta);
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [cursor, debouncedSearch, statusFilter, dateFrom, dateTo, locale]);

  // Re-apply handoff status (?status=processing from command-center cards).
  useEffect(() => {
    setStatusFilter(initialStatus);
    reset();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [initialStatus]);

  // Reset pagination when debounced search changes.
  useEffect(() => {
    reset();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debouncedSearch]);

  useEffect(() => {
    loadOrders();
  }, [loadOrders]);

  const pickPreset = (preset: string) => {
    setDatePreset(preset);
    const range = applyPreset(preset);
    setDateFrom(range.from);
    setDateTo(range.to);
    reset();
  };

  const presets = [
    { id: "all", label: t("orders.dateAllTime") },
    { id: "today", label: t("orders.dateToday") },
    { id: "7d", label: t("orders.dateLast7") },
    { id: "30d", label: t("orders.dateLast30") },
  ];

  return (
    <div className="space-y-6">
      {/* Top back navigation and header */}
      <div className="flex flex-col gap-3">
        <Link
          href="/admin/orders"
          className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
        >
          {t("orders.backToOrders")}
        </Link>

        <PageHeader
          title={t("orders.searchTitle")}
          description={t("orders.searchSubtitle")}
        />
      </div>

      {/* Search & filter card */}
      <div className="rounded-xl border border-border/70 bg-card p-4 sm:p-6 space-y-4">
        <div className="text-sm font-semibold text-foreground">
          {t("orders.searchPageButton")}
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          {/* Search (single unified backend search input) */}
          <div className="sm:col-span-2 space-y-1.5">
            <label htmlFor="ordersSearchQuery" className="text-xs font-medium text-muted-foreground">
              {t("orders.searchLabel")}
            </label>
            <input
              id="ordersSearchQuery"
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={handleSearchKey}
              placeholder={t("orders.searchPlaceholder")}
              className="w-full h-11 px-3 rounded-xl border border-border bg-background text-sm text-foreground focus:outline-hidden focus:border-border hover:bg-muted/30 transition-colors"
            />
          </div>

          {/* Status selector (custom popup) */}
          <div className="space-y-1.5 relative" ref={statusRef}>
            <span id="ordersSearchStatusLabel" className="text-xs font-medium text-muted-foreground">
              {t("orders.filterStatus")}
            </span>
            <button
              type="button"
              onClick={() => setStatusOpen((v) => !v)}
              aria-expanded={statusOpen}
              aria-haspopup="listbox"
              aria-labelledby="ordersSearchStatusLabel"
              className="w-full h-11 px-3 rounded-xl border border-border bg-background text-sm text-foreground text-start hover:bg-muted/30 transition-colors cursor-pointer"
            >
              {statusLabel}
            </button>
            {statusOpen && (
              <div
                role="listbox"
                aria-label={t("orders.filterStatus")}
                className="absolute z-50 mt-2 w-full rounded-xl border border-border/70 bg-card py-1.5 shadow-lg overflow-hidden"
              >
                {statusOptions.map((o) => (
                  <button
                    key={o.value}
                    type="button"
                    role="option"
                    aria-selected={o.value === statusFilter}
                    onClick={() => {
                      setStatusFilter(o.value);
                      setStatusOpen(false);
                      reset();
                    }}
                    className={`w-full px-3 py-2 text-start text-sm transition-colors cursor-pointer ${
                      o.value === statusFilter
                        ? "font-semibold text-foreground bg-muted/40"
                        : "text-muted-foreground hover:text-foreground hover:bg-muted/30"
                    }`}
                  >
                    {o.label}
                  </button>
                ))}
              </div>
            )}
          </div>

        </div>

        {/* Date presets (no native date pickers) */}
        <div className="flex flex-wrap items-center gap-2 pt-2 border-t border-border/50">
          <span className="text-xs text-muted-foreground me-1">{t("orders.filterDateFrom")} / {t("orders.filterDateTo")}:</span>
          {presets.map((p) => (
            <button
              key={p.id}
              type="button"
              onClick={() => pickPreset(p.id)}
              aria-pressed={datePreset === p.id}
              className={`h-9 px-3 rounded-full border text-xs font-medium transition-colors cursor-pointer ${
                datePreset === p.id
                  ? "border-border bg-foreground text-background"
                  : "border-border/70 text-muted-foreground hover:text-foreground hover:bg-muted/30"
              }`}
            >
              {p.label}
            </button>
          ))}
        </div>
      </div>

      {/* Results table */}
      {loadError ? (
        <ErrorState
          title={t("orders.errorTitle")}
          description={t("orders.errorMessage")}
          onRetry={loadOrders}
        />
      ) : (
        <div className="overflow-hidden rounded-xl border border-border/70 bg-card">
          <div className="overflow-x-auto">
            <ResponsiveTable className="w-full text-start text-sm">
              <thead className="text-xs text-muted-foreground bg-muted/30">
                <tr className="border-b border-border/70">
                  <th className="py-3 ps-4 text-start font-medium">{t("orders.colOrderId")}</th>
                  <th className="py-3 text-start font-medium">{t("orders.colCustomer")}</th>
                  <th className="py-3 text-start font-medium">{t("orders.colTotal")}</th>
                  <th className="py-3 text-start font-medium">{t("orders.filterStatus")}</th>
                  <th className="py-3 pe-4 text-end font-medium">{t("orders.colAction")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/50">
                {loading ? (
                  <tr>
                    <td colSpan={5} className="py-12 text-center text-sm text-muted-foreground">
                      <div className="inline-flex items-center gap-2">
                        <div className="size-4 animate-spin rounded-full border-2 border-accent-brand border-t-transparent" />
                        <span>{t("common.loading")}</span>
                      </div>
                    </td>
                  </tr>
                ) : orders.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="py-12 text-center">
                      <p className="text-sm font-medium text-muted-foreground">{t("orders.noOrdersFound")}</p>
                    </td>
                  </tr>
                ) : (
                  orders.map((order) => (
                    <tr key={order.id} className="hover:bg-muted/40 transition-colors">
                      <td className="max-w-36 py-3.5 ps-4">
                        <Link
                          href={`/admin/orders/${order.id}`}
                          className="font-mono text-xs font-semibold text-foreground hover:text-accent-brand transition-colors block truncate"
                        >
                          {order.id}
                        </Link>
                        <div className="text-xs text-muted-foreground">{order.date}</div>
                      </td>
                      <td className="py-3.5">
                        <div className="font-medium text-foreground">{order.customer}</div>
                        <div className="text-xs text-muted-foreground">{order.email}</div>
                      </td>
                      <td className="py-3.5">
                        <div className="font-semibold text-foreground tabular-nums">{order.total}</div>
                      </td>
                      <td className="py-3.5">
                        <StatusPill tone={statusTone(order.status)}>
                          {t(`status.${order.status}`)}
                        </StatusPill>
                      </td>
                      <td className="py-3.5 pe-4 text-end">
                        <Link
                          href={`/admin/orders/${order.id}`}
                          className="text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors"
                        >
                          {t("orders.manage")}
                        </Link>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </ResponsiveTable>
          </div>

          <CursorPager
            hasPrev={hasPrev}
            hasMore={meta.hasMore}
            page={page}
            onPrev={prev}
            onNext={() => next(meta.nextCursor)}
            loading={loading}
            shown={orders.length}
          />
        </div>
      )}
    </div>
  );
}
