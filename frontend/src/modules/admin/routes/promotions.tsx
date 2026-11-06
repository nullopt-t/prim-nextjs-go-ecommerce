import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { useState, useEffect, useMemo, useCallback } from "react";
import { Plus, X } from "lucide-react";
import { adminService } from "@/modules/admin/api/adminService";
import type { Promotion } from "@/modules/promotions/types";
import { useTranslations } from "next-intl";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { formatDateTime } from "@/shared/lib/datetime";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { ConfirmDialog } from "@/shared/ui/confirmDialog";
import { PromotionShareDialog, type PromotionShareContent } from "@/shared/ui/promotionShareDialog";
import { CursorPager, useCursorPagination } from "@/shared/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/shared/lib/pagination";

type SortKey = "code" | "discount" | "usage" | "validity" | "status" | "createdAt";
type SortDir = "asc" | "desc";

interface PromotionView {
  id: string;
  code: string;
  discountType: string;
  value: string;
  discountPercent: number;
  fixedDiscount: number;
  usageLimit: number;
  currentUsage: number;
  status: string;
  startsAt: string;
  expiresAt: string;
  createdAt: string;
}

interface Redemption {
  id: string;
  userName?: string;
  userEmail?: string;
  orderId: string;
  discountAmount: number;
  createdAt: string;
}

function sortValue(promo: PromotionView, key: SortKey): string | number {
  switch (key) {
    case "code":
      return promo.code;
    case "discount":
      // Percentage rows sort by percent; fixed rows by their EGP amount.
      return promo.discountPercent || promo.fixedDiscount || 0;
    case "usage":
      return promo.currentUsage;
    case "validity":
      return promo.expiresAt || "";
    case "status":
      return promo.status;
    case "createdAt":
    default:
      return promo.createdAt || "";
  }
}

function SortButton({ label, active, dir, onClick }: { label: string; active: boolean; dir: SortDir; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`inline-flex items-center gap-1 text-xs font-medium transition-colors hover:text-foreground ${active ? "text-foreground" : "text-muted-foreground"}`}
    >
      {label}
      <span aria-hidden>{!active ? "↕" : dir === "asc" ? "↑" : "↓"}</span>
    </button>
  );
}

export function AdminPromotions() {
  const t = useTranslations("admin");
  const [promotions, setPromotions] = useState<PromotionView[]>([]);
  const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [loadError, setLoadError] = useState(false);
  const [pendingDelete, setPendingDelete] = useState<string | null>(null);
  const [deleting, setDeleting] = useState(false);
  const [sort, setSort] = useState<{ key: SortKey; dir: SortDir }>({ key: "createdAt", dir: "desc" });
  const [logPromo, setLogPromo] = useState<{ id: string; code: string } | null>(null);
  const [redemptions, setRedemptions] = useState<Redemption[]>([]);
  const [logLoading, setLogLoading] = useState(false);
  const [logError, setLogError] = useState(false);
  const [logMeta, setLogMeta] = useState<CursorMeta>(emptyCursorMeta);
  const [sharePromo, setSharePromo] = useState<PromotionShareContent | null>(null);
  const { cursor, hasPrev, next, prev, page } = useCursorPagination();

  const loadPromos = useCallback(async () => {
    try {
      setLoading(true);
      setLoadError(false);
      const res = await adminService.getPromotions({ cursor: cursor || undefined, pageSize: 50 });
      const list = res?.data ?? [];
      setPromotions(list.map((p: Promotion) => ({
        id: p.id,
        code: p.code,
        discountType: p.discountType,
        value: p.discountType === "percentage" ? `${p.discountValue}%` : `EGP ${(p.discountValue / 100).toFixed(2)}`,
        discountPercent: p.discountType === "percentage" ? p.discountValue : 0,
        fixedDiscount: p.discountType === "fixed_amount" ? p.discountValue / 100 : 0,
        usageLimit: p.usageLimit ?? 0,
        currentUsage: p.usageCount ?? 0,
        status: p.status ?? "",
        startsAt: p.startsAt ?? "",
        expiresAt: p.expiresAt ?? "",
        createdAt: p.createdAt ?? ""
      })));
      if (res?.meta) setMeta(res.meta);
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [cursor]);

  useEffect(() => {
    loadPromos();
  }, [loadPromos]);

  const [formData, setFormData] = useState({
    code: "",
    type: "Percentage",
    value: "15",
    usageLimit: 200,
    startDate: "2026-08-01",
    endDate: "2026-12-31"
  });

  const handleToggleStatus = async (id: string, currentStatus: string) => {
    const isActive = currentStatus.toLowerCase() === "active";
    const nextStatus = isActive ? "disabled" : "active";
    try {
      await adminService.updatePromotionStatus(id, nextStatus);
      notify.success(t("promotions.toast.statusUpdated"));
      loadPromos();
    } catch (err) {
      notify.error(apiErrorMessage(err, t("promotions.toast.statusFailed")));
    }
  };

  const handleCreatePromo = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.code) return;

    const isPct = formData.type === "Percentage";
    const numVal = parseFloat(formData.value.replace(/[^0-9.]/g, "")) || 10;
    const discountVal = isPct ? Math.min(100, Math.max(1, Math.round(numVal))) : Math.round(numVal * 100);

    const payload = {
      code: formData.code.toUpperCase().trim(),
      discountType: isPct ? "percentage" : "fixed_amount",
      discountValue: discountVal,
      currency: "EGP",
      minCartAmount: 5000,
      usageLimit: Number(formData.usageLimit) || 100,
      status: "active",
      startsAt: new Date(formData.startDate).toISOString(),
      expiresAt: new Date(formData.endDate).toISOString(),
    };

    try {
      await adminService.createPromotion(payload);
      notify.success(t("promotions.toast.created", { code: payload.code }));
      setIsModalOpen(false);
      loadPromos();
    } catch (err) {
      notify.error(apiErrorMessage(err, t("promotions.toast.createFailed")));
    }
  };

  const handleDelete = (id: string) => {
    setPendingDelete(id);
  };

  const confirmDelete = async () => {
    if (!pendingDelete) return;
    setDeleting(true);
    try {
      await adminService.deletePromotion(pendingDelete);
      notify.success(t("promotions.toast.deleted"));
      setPendingDelete(null);
      loadPromos();
    } catch (err) {
      notify.error(apiErrorMessage(err, t("promotions.toast.deleteFailed")));
    } finally {
      setDeleting(false);
    }
  };

  const openLog = async (promo: { id: string; code: string }) => {
    setLogPromo(promo);
    setRedemptions([]);
    setLogError(false);
    setLogLoading(true);
    try {
      const res = await adminService.getPromotionRedemptions(promo.id, { pageSize: 100 });
      const list: Redemption[] = res?.data ?? [];
      setRedemptions(list);
      setLogMeta((res?.meta ?? emptyCursorMeta) as CursorMeta);
    } catch {
      setLogError(true);
    } finally {
      setLogLoading(false);
    }
  };

  const loadMoreLog = async () => {
    if (!logPromo || !logMeta.hasMore || logLoading) return;
    setLogLoading(true);
    try {
      const res = await adminService.getPromotionRedemptions(logPromo.id, {
        cursor: logMeta.nextCursor,
        pageSize: 100,
      });
      const list: Redemption[] = res?.data ?? [];
      setRedemptions((prev) => [...prev, ...list]);
      setLogMeta((res?.meta ?? emptyCursorMeta) as CursorMeta);
    } catch {
      setLogError(true);
    } finally {
      setLogLoading(false);
    }
  };

  const openShare = (promo: PromotionView) => {
    setSharePromo({
      code: promo.code,
      value: promo.value,
      typeLabel:
        promo.discountType === "percentage"
          ? t("promotions.type.percentage")
          : t("promotions.type.fixedAmount"),
      validityLabel: t("promotions.share.validUntil", {
        date: formatDateTime(promo.expiresAt) || "—",
      }),
    });
  };

  const sorted = useMemo(() => {
    const dir = sort.dir === "asc" ? 1 : -1;
    return [...promotions].sort((a: PromotionView, b: PromotionView) => {
      const av = sortValue(a, sort.key);
      const bv = sortValue(b, sort.key);
      const cmp =
        typeof av === "number" && typeof bv === "number"
          ? av - bv
          : String(av).localeCompare(String(bv));
      return cmp * dir;
    });
  }, [promotions, sort]);

  const toggleSort = (key: SortKey) => {
    setSort((prev) =>
      prev.key === key ? { key, dir: prev.dir === "asc" ? "desc" : "asc" } : { key, dir: "asc" }
    );
  };

  if (loadError) {
    return <ErrorState onRetry={loadPromos} />;
  }

  return (
    <div className="space-y-5">
      <PageHeader
        title={t("promotions.title")}
        actions={
          <button
            onClick={() => setIsModalOpen(true)}
            className="inline-flex items-center gap-1.5 rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
          >
            <Plus className="size-4" />
            {t("promotions.create")}
          </button>
        }
      />

      <div className="overflow-hidden rounded-lg border border-border bg-background">
        <div className="overflow-x-auto">
          <ResponsiveTable className="w-full text-start text-sm md:min-w-max">
            <thead className="text-xs text-muted-foreground">
              <tr className="border-b border-border">
                <th className="py-2.5 ps-4 text-start font-medium">
                  <SortButton label={t("promotions.colCode")} active={sort.key === "code"} dir={sort.dir} onClick={() => toggleSort("code")} />
                </th>
                <th className="py-2.5 text-start font-medium">
                  <SortButton label={t("promotions.colDiscount")} active={sort.key === "discount"} dir={sort.dir} onClick={() => toggleSort("discount")} />
                </th>
                <th className="py-2.5 text-start font-medium">
                  <SortButton label={t("promotions.colUsage")} active={sort.key === "usage"} dir={sort.dir} onClick={() => toggleSort("usage")} />
                </th>
                <th className="py-2.5 text-start font-medium">
                  <SortButton label={t("promotions.colValidity")} active={sort.key === "validity"} dir={sort.dir} onClick={() => toggleSort("validity")} />
                </th>
                <th className="py-2.5 text-start font-medium">
                  <SortButton label={t("promotions.colStatus")} active={sort.key === "status"} dir={sort.dir} onClick={() => toggleSort("status")} />
                </th>
                <th className="py-2.5 pe-4 text-end font-medium">{t("promotions.colActions")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {sorted.map((promo) => (
                <tr key={promo.id} className="transition-colors hover:bg-secondary/30">
                  <td className="py-3 ps-4">
                    <span className="font-mono text-xs font-medium tracking-wider text-foreground">
                      {promo.code}
                    </span>
                  </td>
                  <td className="py-3">
                    <div className="font-medium text-foreground">{promo.value}</div>
                    <div className="mt-0.5 text-xs text-muted-foreground">
                      {t("promotions.typeLabel", {
                        type:
                          promo.discountType === "percentage"
                            ? t("promotions.type.percentage")
                            : t("promotions.type.fixedAmount"),
                      })}
                    </div>
                  </td>
                  <td className="py-3 text-xs text-muted-foreground">
                    {t("promotions.redeemed", { used: promo.currentUsage, limit: promo.usageLimit })}
                  </td>
                  <td className="py-3 text-xs text-muted-foreground">
                    <div>{t("promotions.valid", { date: formatDateTime(promo.startsAt) || "—" })}</div>
                    <div className="mt-0.5">{t("promotions.expires", { date: formatDateTime(promo.expiresAt) || "—" })}</div>
                  </td>
                  <td className="py-3">
                    <StatusPill tone={statusTone(promo.status)}>
                      {t(`status.${String(promo.status).toLowerCase()}`) || promo.status}
                    </StatusPill>
                  </td>
                  <td className="py-3 pe-4 text-end">
                    <div className="inline-flex items-center gap-1">
                      <button
                        onClick={() => handleToggleStatus(promo.id, promo.status)}
                        className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                      >
                        {promo.status.toLowerCase() === "active"
                          ? t("promotions.actions.deactivate")
                          : t("promotions.actions.activate")}
                      </button>
                      <button
                        onClick={() => openShare(promo)}
                        className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                      >
                        {t("promotions.share.button")}
                      </button>
                      <button
                        onClick={() => openLog({ id: promo.id, code: promo.code })}
                        className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                      >
                        {t("promotions.viewLog")}
                      </button>
                      <button
                        onClick={() => handleDelete(promo.id)}
                        className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                      >
                        {t("actions.delete")}
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
              {sorted.length === 0 ? (
                <tr>
                  <td colSpan={6} className="py-10 text-center text-sm text-muted-foreground">
                    {t("promotions.empty")}
                  </td>
                </tr>
              ) : null}
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
          shown={sorted.length}
        />
      </div>

      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="w-full max-w-md space-y-4 rounded-lg border border-border bg-background p-5 shadow-xl">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <h3 className="text-sm font-semibold text-foreground">{t("promotions.create")}</h3>
              <button
                onClick={() => setIsModalOpen(false)}
                className="text-muted-foreground hover:text-foreground"
                aria-label={t("actions.cancel")}
              >
                <X className="size-4" />
              </button>
            </div>

            <form onSubmit={handleCreatePromo} className="space-y-4">
              <div className="space-y-1">
                <label className="text-sm font-medium text-foreground">{t("promotions.form.code")}</label>
                <input
                  type="text"
                  required
                  placeholder={t("promotions.form.codePlaceholder")}
                  value={formData.code}
                  onChange={(e) => setFormData({ ...formData, code: e.target.value })}
                  className="w-full rounded-md border border-border bg-background px-3 py-2 font-mono text-sm uppercase text-foreground"
                />
              </div>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="space-y-1">
                  <label className="text-sm font-medium text-foreground">{t("promotions.form.discountType")}</label>
                  <select
                    value={formData.type}
                    onChange={(e) => setFormData({ ...formData, type: e.target.value })}
                    className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
                  >
                    <option value="Percentage">{t("promotions.form.typePercentage")}</option>
                    <option value="Fixed Amount">{t("promotions.form.typeFixed")}</option>
                    <option value="Free Shipping">{t("promotions.form.typeFreeShipping")}</option>
                  </select>
                </div>
                <div className="space-y-1">
                  <label className="text-sm font-medium text-foreground">{t("promotions.form.valueLabel")}</label>
                  <input
                    type="text"
                    required
                    placeholder={t("promotions.form.valuePlaceholder")}
                    value={formData.value}
                    onChange={(e) => setFormData({ ...formData, value: e.target.value })}
                    className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
                  />
                </div>
              </div>

              <div className="space-y-1">
                <label className="text-sm font-medium text-foreground">{t("promotions.form.usageLimit")}</label>
                <input
                  type="number"
                  value={formData.usageLimit}
                  onChange={(e) => setFormData({ ...formData, usageLimit: Number(e.target.value) || 0 })}
                  className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
                />
              </div>

              <div className="flex justify-end gap-2 border-t border-border pt-3">
                <button type="button" onClick={() => setIsModalOpen(false)} className="rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground hover:bg-secondary">{t("actions.cancel")}</button>
                <button type="submit" className="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground">{t("promotions.form.submit")}</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {logPromo && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="flex max-h-[80vh] w-full max-w-2xl flex-col space-y-4 rounded-lg border border-border bg-background p-5 shadow-xl">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <h3 className="text-sm font-semibold text-foreground">
                {t("promotions.log.title", { code: logPromo.code })}
              </h3>
              <button
                onClick={() => setLogPromo(null)}
                className="text-muted-foreground hover:text-foreground"
                aria-label={t("actions.close")}
              >
                <X className="size-4" />
              </button>
            </div>

            <div className="overflow-y-auto">
              {logLoading ? (
                <div className="py-10 text-center text-sm text-muted-foreground">
                  {t("promotions.log.loading")}
                </div>
              ) : logError ? (
                <p className="py-10 text-center text-sm text-destructive">
                  {t("promotions.log.loadFailed")}
                </p>
              ) : redemptions.length === 0 ? (
                <p className="py-10 text-center text-sm text-muted-foreground">
                  {t("promotions.log.empty")}
                </p>
              ) : (
                <ResponsiveTable className="w-full text-start text-sm md:min-w-max">
                  <thead className="text-xs text-muted-foreground">
                    <tr className="border-b border-border">
                      <th className="py-2.5 text-start font-medium">{t("promotions.log.colUser")}</th>
                      <th className="py-2.5 text-start font-medium">{t("promotions.log.colOrder")}</th>
                      <th className="py-2.5 text-start font-medium">{t("promotions.log.colDiscount")}</th>
                      <th className="py-2.5 text-start font-medium">{t("promotions.log.colDate")}</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border">
                    {redemptions.map((r) => (
                      <tr key={r.id} className="transition-colors hover:bg-secondary/30">
                        <td className="py-3">
                          <div className="font-medium text-foreground">
                            {r.userName || t("promotions.log.customer")}
                          </div>
                          <div className="text-xs text-muted-foreground">
                            {r.userEmail || t("promotions.log.unknownUser")}
                          </div>
                        </td>
                        <td className="py-3 font-mono text-xs text-muted-foreground" title={r.orderId}>
                          {String(r.orderId).slice(0, 8)}
                        </td>
                        <td className="py-3 font-medium text-foreground tabular-nums">
                          EGP {(r.discountAmount / 100).toFixed(2)}
                        </td>
                        <td className="py-3 text-xs text-muted-foreground">
                          {formatDateTime(r.createdAt) || "—"}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </ResponsiveTable>
              )}
            </div>

            {!logLoading && logMeta.hasMore ? (
              <div className="flex justify-center border-t border-border pt-3">
                <button
                  onClick={loadMoreLog}
                  className="rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground hover:bg-secondary"
                >
                  {t("promotions.log.loadMore")}
                </button>
              </div>
            ) : null}
          </div>
        </div>
      )}

      <PromotionShareDialog
        isOpen={sharePromo !== null}
        content={sharePromo}
        onClose={() => setSharePromo(null)}
      />

      <ConfirmDialog
        isOpen={pendingDelete !== null}
        title={t("promotions.confirmDelete")}
        confirmLabel={t("actions.delete")}
        cancelLabel={t("actions.cancel")}
        loading={deleting}
        onConfirm={confirmDelete}
        onClose={() => setPendingDelete(null)}
      />
    </div>
  );
}
