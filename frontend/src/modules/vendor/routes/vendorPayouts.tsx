import { useState, useEffect } from "react";

import { DollarSign, Landmark, CheckCircle2, Clock, X } from "lucide-react";
import {
  INITIAL_VENDOR_PAYOUTS,
  getLocalVendorData,
  setLocalVendorData
} from "../data/mockVendorData";
import { vendorService } from "@/modules/vendor/api/vendorService";
import { notify } from "@/shared/lib/notify";
import { useTranslations } from "next-intl";

import { useVendorContext } from "./vendorContext";

export function VendorPayouts() {
  const t = useTranslations("vendor");
  const { vendor, refreshVendors } = useVendorContext();

  const [payouts, setPayouts] = useState(() => getLocalVendorData("payouts", INITIAL_VENDOR_PAYOUTS));
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [availableBalance, setAvailableBalance] = useState(vendor?.balance !== undefined ? vendor.balance : 521.96);
  const [pendingBalance, setPendingBalance] = useState(179.99);
  const [payoutAmount, setPayoutAmount] = useState(
    (vendor?.balance !== undefined ? vendor.balance : 521.96).toFixed(2)
  );

  useEffect(() => {
    async function loadPayouts() {
      try {
        const [payoutsRes, statsRes] = await Promise.allSettled([
          vendorService.getPayouts(vendor?.id),
          vendorService.getVendorStats(vendor?.id)
        ]);

        if (payoutsRes.status === "fulfilled" && payoutsRes.value) {
          const list = Array.isArray(payoutsRes.value) ? payoutsRes.value : Array.isArray(payoutsRes.value?.data) ? payoutsRes.value.data : [];
          if (list.length > 0) {
            setPayouts(list);
            setLocalVendorData("payouts", list);
          }
        }

        if (statsRes.status === "fulfilled" && statsRes.value) {
          const stats = statsRes.value?.data || statsRes.value;
          if (stats && stats.availableBalance !== undefined) {
            setAvailableBalance(stats.availableBalance);
            setPayoutAmount(stats.availableBalance.toFixed(2));
          }
          if (stats && stats.pendingBalance !== undefined) {
            setPendingBalance(stats.pendingBalance);
          }
        } else if (vendor?.balance !== undefined) {
          setAvailableBalance(vendor.balance);
          setPayoutAmount(vendor.balance.toFixed(2));
        }
      } catch {
        // Retain local state
      }
    }
    loadPayouts();
  }, [vendor?.id, vendor?.balance]);

  const savePayouts = (updated: any[]) => {
    setPayouts(updated);
    setLocalVendorData("payouts", updated);
  };

  const statusLabel = (status: string) => {
    const map: Record<string, string> = {
      Processing: "status.processing",
      Completed: "status.completed",
      Pending: "status.pending",
      Paid: "status.paid"
    };
    return map[status] ? t(map[status]) : status;
  };

  const lifetimePaidOut = payouts.reduce((acc, p) => {
    const val = typeof p.amount === "number" ? p.amount : parseFloat(String(p.amount).replace(/[^0-9.-]+/g, "")) || 0;
    return acc + val;
  }, 0);

  const handleRequestPayout = async (e: React.FormEvent) => {
    e.preventDefault();
    const amountVal = parseFloat(payoutAmount);
    if (!amountVal || amountVal <= 0 || amountVal > availableBalance) {
      notify.error(t("toasts.invalidPayoutAmount"));
      return;
    }

    const newPayout = {
      id: `PO-${Math.floor(Math.random() * 9000 + 1000)}`,
      vendorId: vendor?.id || "v-1",
      amount: `EGP ${amountVal.toFixed(2)}`,
      method: "Direct Deposit (Chase ending in 8812)",
      status: "Processing",
      requestedDate: new Date().toISOString().split("T")[0],
      processedDate: "Pending Clearance",
      reference: `TRX-ACH-${Math.floor(Math.random() * 900000 + 100000)}`
    };

    const updatedList = [newPayout, ...payouts];
    savePayouts(updatedList);
    setAvailableBalance((prev) => Math.max(0, prev - amountVal));

    try {
      const res = await vendorService.requestPayout({
        vendorId: vendor?.id || "v-1",
        amount: amountVal,
        method: "Direct Deposit (Chase ending in 8812)"
      });
      if (res?.data) {
        refreshVendors?.();
      }
    } catch {
      // Mock fallback
    }

    setIsModalOpen(false);
    notify.success(t("toasts.payoutSubmitted", { amount: amountVal.toFixed(2) }));
  };

  return (
    <div className="space-y-6 w-full">
      {/* Top Banner */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t("payouts.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {t("payouts.subtitle")}
          </p>
        </div>

        <button
          onClick={() => setIsModalOpen(true)}
          className="flex items-center gap-2 px-5 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all shrink-0"
        >
          <DollarSign className="size-4" />
          <span>{t("actions.requestPayout")}</span>
        </button>
      </div>

      {/* Financial Summary Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-2">
          <div className="flex items-center justify-between text-muted-foreground text-xs font-semibold">
            <span>{t("payouts.availableBalance")}</span>
            <CheckCircle2 className="size-4 text-emerald-500" />
          </div>
          <div className="text-3xl font-bold text-foreground">EGP {availableBalance.toFixed(2)}</div>
          <p className="text-[11px] text-muted-foreground">{t("payouts.availableBalanceNote")}</p>
        </div>

        <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-2">
          <div className="flex items-center justify-between text-muted-foreground text-xs font-semibold">
            <span>{t("payouts.pendingClearance")}</span>
            <Clock className="size-4 text-amber-500" />
          </div>
          <div className="text-3xl font-bold text-foreground">EGP {pendingBalance.toFixed(2)}</div>
          <p className="text-[11px] text-muted-foreground">{t("payouts.pendingClearanceNote")}</p>
        </div>

        <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-2">
          <div className="flex items-center justify-between text-muted-foreground text-xs font-semibold">
            <span>{t("payouts.lifetimePaidOut")}</span>
            <Landmark className="size-4 text-accent-brand" />
          </div>
          <div className="text-3xl font-bold text-foreground">EGP {lifetimePaidOut.toFixed(2)}</div>
          <p className="text-[11px] text-muted-foreground">{t("payouts.lifetimePaidOutNote")}</p>
        </div>
      </div>

      {/* Linked Bank Account Card */}
      <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-3">
        <div className="flex items-center justify-between border-b border-border pb-3">
          <div className="flex items-center gap-3">
            <div className="size-10 rounded-xl bg-secondary flex items-center justify-center text-foreground font-bold">
              <Landmark className="size-5 text-accent-brand" />
            </div>
            <div>
              <div className="font-bold text-sm text-foreground">Chase Business Checking</div>
              <div className="text-xs text-muted-foreground">{t("payouts.accountEnding")}</div>
            </div>
          </div>
          <span className="px-3 py-1 bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-400 text-xs font-bold rounded-full">
            {t("payouts.verifiedAccount")}
          </span>
        </div>

        <div className="flex items-center justify-between text-xs text-muted-foreground">
          <span>{t("payouts.schedule")} <strong>{t("payouts.scheduleValue")}</strong></span>
          <span>{t("payouts.commissionFee")} <strong>{vendor?.commissionRate || 10}%</strong></span>
        </div>
      </div>

      {/* Historical Payouts Table */}
      <div className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs space-y-4">
        <div className="p-6 border-b border-border">
          <h3 className="font-bold text-base text-foreground">{t("payouts.historyTitle")}</h3>
          <p className="text-xs text-muted-foreground mt-0.5">{t("payouts.historySubtitle")}</p>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full min-w-max text-left text-xs">
            <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
              <tr>
                <th className="py-3.5 pl-6">{t("payouts.table.payoutId")}</th>
                <th className="py-3.5">{t("payouts.table.amount")}</th>
                <th className="py-3.5">{t("payouts.table.method")}</th>
                <th className="py-3.5">{t("payouts.table.requestedDate")}</th>
                <th className="py-3.5">{t("payouts.table.settlementDate")}</th>
                <th className="py-3.5">{t("payouts.table.reference")}</th>
                <th className="py-3.5 pr-6 text-right">{t("payouts.table.status")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {payouts.map((p: any) => (
                <tr key={p.id} className="hover:bg-secondary/20 transition-colors">
                  <td className="py-3.5 pl-6 font-mono font-bold text-foreground">{p.id}</td>
                  <td className="py-3.5 font-bold text-foreground text-sm">{p.amount}</td>
                  <td className="py-3.5 text-muted-foreground">{p.method}</td>
                  <td className="py-3.5 text-muted-foreground">{p.requestedDate}</td>
                  <td className="py-3.5 text-muted-foreground">{p.processedDate}</td>
                  <td className="py-3.5 font-mono text-muted-foreground">{p.reference}</td>
                  <td className="py-3.5 pr-6 text-right">
                    <span
                      className={`px-2.5 py-0.5 rounded-full text-[10px] font-bold ${
                        p.status === "Completed"
                          ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-400"
                          : "bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-400"
                      }`}
                    >
                      {statusLabel(p.status)}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Request Payout Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-background border border-border rounded-2xl max-w-md w-full p-6 shadow-xl space-y-4">
            <div className="flex justify-between items-center border-b border-border pb-3">
              <h3 className="font-bold text-base text-foreground">{t("payouts.modalTitle")}</h3>
              <button onClick={() => setIsModalOpen(false)} className="text-muted-foreground">
                <X className="size-5" />
              </button>
            </div>

            <form onSubmit={handleRequestPayout} className="space-y-4">
              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">{t("payouts.withdrawalAmount")}</label>
                <input
                  type="number"
                  step="0.01"
                  max={availableBalance}
                  required
                  value={payoutAmount}
                  onChange={(e) => setPayoutAmount(e.target.value)}
                  className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-base font-bold text-foreground"
                />
                <div className="text-[11px] text-muted-foreground mt-1">
                  {t("payouts.maximumWithdrawable")} <strong>EGP {availableBalance.toFixed(2)}</strong>
                </div>
              </div>

              <div className="p-3 bg-secondary/30 rounded-xl border border-border text-xs space-y-1">
                <div className="font-semibold text-foreground">{t("payouts.destinationAccount")}</div>
                <div className="text-muted-foreground">{t("payouts.destinationAccountValue")}</div>
                <div className="text-[10px] text-emerald-600 dark:text-emerald-400 font-medium">{t("payouts.estimatedArrival")}</div>
              </div>

              <div className="flex justify-end gap-3 pt-3 border-t border-border">
                <button type="button" onClick={() => setIsModalOpen(false)} className="px-4 py-2 text-xs border border-border rounded-xl">
                  {t("actions.cancel")}
                </button>
                <button type="submit" className="px-4 py-2 text-xs bg-accent-brand text-white font-semibold rounded-xl">
                  {t("actions.confirmTransfer")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
