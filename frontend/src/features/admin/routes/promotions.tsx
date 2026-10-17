import { useState, useEffect, useMemo, useCallback } from "react";
import { Plus, Trash2, X, ArrowUp, ArrowDown, ArrowUpDown, ScrollText, Loader2, Share2 } from "lucide-react";
import { adminService } from "@/services/admin";
import { useTranslations } from "next-intl";
import { notify, apiErrorMessage } from "@/lib/notify";
import { formatDateTime } from "@/lib/datetime";
import { ErrorState } from "@/components/ui/errorState";
import { ConfirmDialog } from "@/components/ui/confirmDialog";
import { PromotionShareDialog, type PromotionShareContent } from "@/components/ui/promotionShareDialog";
import { CursorPager, useCursorPagination } from "@/components/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/lib/pagination";

type SortKey = "code" | "discount" | "usage" | "validity" | "status" | "createdAt";
type SortDir = "asc" | "desc";

function sortValue(promo: any, key: SortKey): string | number {
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
  const Icon = !active ? ArrowUpDown : dir === "asc" ? ArrowUp : ArrowDown;
  return (
    <button
      type="button"
      onClick={onClick}
      className={`inline-flex items-center gap-1 transition-colors hover:text-foreground ${active ? "text-foreground" : ""}`}
    >
      {label}
      <Icon className="size-3" />
    </button>
  );
}

export function AdminPromotions() {
  const t = useTranslations("admin");
  const [promotions, setPromotions] = useState<any[]>([]);
  const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [loadError, setLoadError] = useState(false);
  const [pendingDelete, setPendingDelete] = useState<string | null>(null);
  const [deleting, setDeleting] = useState(false);
  const [sort, setSort] = useState<{ key: SortKey; dir: SortDir }>({ key: "createdAt", dir: "desc" });
  const [logPromo, setLogPromo] = useState<{ id: string; code: string } | null>(null);
  const [redemptions, setRedemptions] = useState<any[]>([]);
  const [logLoading, setLogLoading] = useState(false);
  const [logError, setLogError] = useState(false);
  const [logMeta, setLogMeta] = useState<CursorMeta>(emptyCursorMeta);
  const [sharePromo, setSharePromo] = useState<PromotionShareContent | null>(null);
  const { cursor, hasPrev, next, prev, page } = useCursorPagination();

  const loadPromos = useCallback(async () => {
    try {
      setLoading(true);
      setLoadError(false);
      const res: any = await adminService.getPromotions({ cursor: cursor || undefined, pageSize: 50 });
      const list = Array.isArray(res?.data) ? res.data : (Array.isArray(res) ? res : []);
      setPromotions(list.map((p: any) => ({
        id: p.id,
        code: p.code,
        discountType: p.discountType,
        value: p.discountType === "percentage" ? `${p.discountValue}%` : `EGP ${(p.discountValue / 100).toFixed(2)}`,
        discountPercent: p.discountType === "percentage" ? p.discountValue : 0,
        fixedDiscount: p.discountType === "fixed_amount" ? p.discountValue / 100 : 0,
        usageLimit: p.usageLimit || 500,
        currentUsage: p.usageCount ?? 0,
        status: p.status ? p.status.charAt(0).toUpperCase() + p.status.slice(1) : "Active",
        startsAt: p.startsAt || "",
        expiresAt: p.expiresAt || "",
        createdAt: p.createdAt || ""
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
    const nextStatus = currentStatus === "Active" ? "disabled" : "active";
    try {
      await adminService.updatePromotionStatus(id, nextStatus);
      notify.success(t("promotions.toast.statusUpdated"));
      loadPromos();
    } catch (err: any) {
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
    } catch (err: any) {
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
    } catch (err: any) {
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
      const res: any = await adminService.getPromotionRedemptions(promo.id, { pageSize: 100 });
      const list = Array.isArray(res?.data) ? res.data : (Array.isArray(res) ? res : []);
      setRedemptions(list);
      setLogMeta(res?.meta ?? emptyCursorMeta);
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
      const res: any = await adminService.getPromotionRedemptions(logPromo.id, {
        cursor: logMeta.nextCursor,
        pageSize: 100,
      });
      const list = Array.isArray(res?.data) ? res.data : (Array.isArray(res) ? res : []);
      setRedemptions((prev) => [...prev, ...list]);
      setLogMeta(res?.meta ?? emptyCursorMeta);
    } catch {
      setLogError(true);
    } finally {
      setLogLoading(false);
    }
  };

  const openShare = (promo: any) => {
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
    return [...promotions].sort((a: any, b: any) => {
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
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t("promotions.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {t("promotions.subtitle")}
          </p>
        </div>
        <button
          onClick={() => setIsModalOpen(true)}
          className="flex items-center gap-2 px-4 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all shrink-0"
        >
          <Plus className="size-4" />
          <span>{t("promotions.create")}</span>
        </button>
      </div>

      <div className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs">
        <div className="overflow-x-auto">
          <table className="w-full text-start text-xs">
            <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
              <tr>
                <th className="py-3.5 ps-4">
                  <SortButton label={t("promotions.colCode")} active={sort.key === "code"} dir={sort.dir} onClick={() => toggleSort("code")} />
                </th>
                <th className="py-3.5">
                  <SortButton label={t("promotions.colDiscount")} active={sort.key === "discount"} dir={sort.dir} onClick={() => toggleSort("discount")} />
                </th>
                <th className="py-3.5">
                  <SortButton label={t("promotions.colUsage")} active={sort.key === "usage"} dir={sort.dir} onClick={() => toggleSort("usage")} />
                </th>
                <th className="py-3.5">
                  <SortButton label={t("promotions.colValidity")} active={sort.key === "validity"} dir={sort.dir} onClick={() => toggleSort("validity")} />
                </th>
                <th className="py-3.5">
                  <SortButton label={t("promotions.colStatus")} active={sort.key === "status"} dir={sort.dir} onClick={() => toggleSort("status")} />
                </th>
                <th className="py-3.5 pe-4 text-end">{t("promotions.colActions")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {sorted.map((promo) => (
                <tr key={promo.id} className="hover:bg-secondary/20 transition-colors">
                  <td className="py-3.5 ps-4">
                    <span className="font-mono font-bold text-sm text-accent-brand tracking-wider bg-accent-brand/10 px-2.5 py-1 rounded-lg border border-accent-brand/20">
                      {promo.code}
                    </span>
                  </td>
                  <td className="py-3.5">
                    <div className="font-bold text-foreground">{promo.value}</div>
                    <div className="text-[10px] text-muted-foreground mt-0.5">
                      {t("promotions.typeLabel", {
                        type:
                          promo.discountType === "percentage"
                            ? t("promotions.type.percentage")
                            : t("promotions.type.fixedAmount"),
                      })}
                    </div>
                  </td>
                  <td className="py-3.5 text-muted-foreground">
                    {t("promotions.redeemed", { used: promo.currentUsage, limit: promo.usageLimit })}
                  </td>
                  <td className="py-3.5 text-muted-foreground">
                    <div>{t("promotions.valid", { date: formatDateTime(promo.startsAt) || "—" })}</div>
                    <div className="text-[10px] mt-0.5">{t("promotions.expires", { date: formatDateTime(promo.expiresAt) || "—" })}</div>
                  </td>
                  <td className="py-3.5">
                    <button
                      onClick={() => handleToggleStatus(promo.id, promo.status)}
                      className="text-foreground hover:text-accent-brand transition-colors"
                    >
                      {t(`status.${String(promo.status).toLowerCase()}`) || promo.status}
                    </button>
                  </td>
                  <td className="py-3.5 pe-4 text-end">
                    <div className="flex items-center justify-end gap-1">
                      <button
                        onClick={() => openShare(promo)}
                        className="p-1.5 text-muted-foreground hover:text-accent-brand transition-colors"
                        title={t("promotions.share.button")}
                      >
                        <Share2 className="size-4" />
                      </button>
                      <button
                        onClick={() => openLog({ id: promo.id, code: promo.code })}
                        className="p-1.5 text-muted-foreground hover:text-accent-brand transition-colors"
                        title={t("promotions.viewLog")}
                      >
                        <ScrollText className="size-4" />
                      </button>
                      <button
                        onClick={() => handleDelete(promo.id)}
                        className="p-1.5 text-muted-foreground hover:text-destructive transition-colors"
                        title={t("actions.delete")}
                      >
                        <Trash2 className="size-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
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
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-background border border-border rounded-2xl max-w-md w-full p-6 shadow-xl space-y-4">
            <div className="flex justify-between items-center border-b border-border pb-3">
              <h3 className="font-bold text-base text-foreground">{t("promotions.create")}</h3>
              <button onClick={() => setIsModalOpen(false)} className="text-muted-foreground">
                <X className="size-5" />
              </button>
            </div>

            <form onSubmit={handleCreatePromo} className="space-y-4">
              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">{t("promotions.form.code")}</label>
                <input
                  type="text"
                  required
                  placeholder={t("promotions.form.codePlaceholder")}
                  value={formData.code}
                  onChange={(e) => setFormData({ ...formData, code: e.target.value })}
                  className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs font-mono uppercase text-foreground"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-foreground">{t("promotions.form.discountType")}</label>
                  <select
                    value={formData.type}
                    onChange={(e) => setFormData({ ...formData, type: e.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                  >
                    <option value="Percentage">{t("promotions.form.typePercentage")}</option>
                    <option value="Fixed Amount">{t("promotions.form.typeFixed")}</option>
                    <option value="Free Shipping">{t("promotions.form.typeFreeShipping")}</option>
                  </select>
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-foreground">{t("promotions.form.valueLabel")}</label>
                  <input
                    type="text"
                    required
                    placeholder={t("promotions.form.valuePlaceholder")}
                    value={formData.value}
                    onChange={(e) => setFormData({ ...formData, value: e.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                  />
                </div>
              </div>

              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">{t("promotions.form.usageLimit")}</label>
                <input
                  type="number"
                  value={formData.usageLimit}
                  onChange={(e) => setFormData({ ...formData, usageLimit: Number(e.target.value) || 0 })}
                  className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                />
              </div>

              <div className="flex justify-end gap-3 pt-3 border-t border-border">
                <button type="button" onClick={() => setIsModalOpen(false)} className="px-4 py-2 text-xs border border-border rounded-xl">{t("actions.cancel")}</button>
                <button type="submit" className="px-4 py-2 text-xs bg-accent-brand text-white rounded-xl">{t("promotions.form.submit")}</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {logPromo && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-background border border-border rounded-2xl max-w-2xl w-full p-6 shadow-xl space-y-4 flex flex-col max-h-[80vh]">
            <div className="flex justify-between items-center border-b border-border pb-3">
              <h3 className="font-bold text-base text-foreground">
                {t("promotions.log.title", { code: logPromo.code })}
              </h3>
              <button onClick={() => setLogPromo(null)} className="text-muted-foreground">
                <X className="size-5" />
              </button>
            </div>

            <div className="overflow-y-auto">
              {logLoading ? (
                <div className="flex items-center justify-center py-10 text-muted-foreground">
                  <Loader2 className="size-5 animate-spin" />
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
                <table className="w-full text-start text-xs">
                  <thead className="text-muted-foreground font-semibold uppercase tracking-wider border-b border-border">
                    <tr>
                      <th className="py-2.5">{t("promotions.log.colUser")}</th>
                      <th className="py-2.5">{t("promotions.log.colOrder")}</th>
                      <th className="py-2.5">{t("promotions.log.colDiscount")}</th>
                      <th className="py-2.5">{t("promotions.log.colDate")}</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border/60">
                    {redemptions.map((r) => (
                      <tr key={r.id}>
                        <td className="py-2.5">
                          <div className="font-medium text-foreground">
                            {r.userName || t("promotions.log.customer")}
                          </div>
                          <div className="text-[10px] text-muted-foreground">
                            {r.userEmail || t("promotions.log.unknownUser")}
                          </div>
                        </td>
                        <td className="py-2.5 font-mono text-[10px] text-muted-foreground" title={r.orderId}>
                          {String(r.orderId).slice(0, 8)}
                        </td>
                        <td className="py-2.5 text-foreground">
                          EGP {(r.discountAmount / 100).toFixed(2)}
                        </td>
                        <td className="py-2.5 text-muted-foreground">
                          {formatDateTime(r.createdAt)}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>

            {!logLoading && logMeta.hasMore && (
              <div className="flex justify-center pt-2 border-t border-border">
                <button
                  onClick={loadMoreLog}
                  className="px-4 py-2 text-xs font-semibold bg-secondary hover:bg-secondary/80 text-foreground border border-border rounded-xl"
                >
                  {t("promotions.log.loadMore")}
                </button>
              </div>
            )}
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
