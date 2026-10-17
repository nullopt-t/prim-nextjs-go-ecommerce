"use client";

import React, { useState } from "react";
import { 
  X, 
  UserX, 
  AlertOctagon, 
  Clock, 
  CheckCircle2,
  Loader2 
} from "lucide-react";
import { AdminUser } from "@/services/admin";
import { useTranslations } from "next-intl";

interface SuspendUserModalProps {
  user: AdminUser;
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (userId: string, status: AdminUser["status"], suspendedUntil: string | null) => Promise<void>;
}

export function SuspendUserModal({ user, isOpen, onClose, onConfirm }: SuspendUserModalProps) {
  const t = useTranslations("admin");
  const isCurrentlySuspended = user.status === "suspended";
  const [duration, setDuration] = useState<"24h" | "7d" | "30d" | "indefinite">("7d");
  const [reason, setReason] = useState("");
  const [submitting, setSubmitting] = useState(false);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      if (isCurrentlySuspended) {
        // Reactivating user
        await onConfirm(user.id, "active", null);
      } else {
        // Suspending user
        let untilDate: Date | null = null;
        const now = new Date();
        if (duration === "24h") {
          untilDate = new Date(now.getTime() + 24 * 60 * 60 * 1000);
        } else if (duration === "7d") {
          untilDate = new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000);
        } else if (duration === "30d") {
          untilDate = new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000);
        }

        await onConfirm(user.id, "suspended", untilDate ? untilDate.toISOString() : null);
      }
      onClose();
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs animate-in fade-in duration-150">
      <div 
        className="bg-card border border-border w-full max-w-md rounded-2xl shadow-2xl overflow-hidden flex flex-col"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className={`px-6 py-5 border-b border-border flex items-center justify-between ${
          isCurrentlySuspended ? "bg-emerald-500/10" : "bg-rose-500/10"
        }`}>
          <div className="flex items-center gap-3">
            <div className={`p-2.5 rounded-xl border ${
              isCurrentlySuspended 
                ? "bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 border-emerald-500/30"
                : "bg-rose-500/20 text-rose-600 dark:text-rose-400 border-rose-500/30"
            }`}>
              {isCurrentlySuspended ? <CheckCircle2 className="size-5" /> : <UserX className="size-5" />}
            </div>
            <div>
              <h2 className="text-base font-bold text-foreground">
                {isCurrentlySuspended ? t("modals.suspend.reactivateTitle") : t("modals.suspend.suspendTitle")}
              </h2>
              <p className="text-xs text-muted-foreground">
                {isCurrentlySuspended ? t("modals.suspend.reactivateSubtitle") : t("modals.suspend.suspendSubtitle")}
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors"
          >
            <X className="size-4" />
          </button>
        </div>

        {/* User Snapshot */}
        <div className="px-6 py-3.5 bg-secondary/30 border-b border-border text-xs flex items-center justify-between">
          <div className="flex flex-col">
            <span className="font-semibold text-foreground">{user.fullName || t("customers.unnamed")}</span>
            <span className="font-mono text-muted-foreground">{user.email}</span>
          </div>
          <span className={`px-2.5 py-1 rounded-md text-[11px] font-bold uppercase tracking-wider border ${
            user.status === "active"
              ? "bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 border-emerald-500/30"
              : "bg-rose-500/15 text-rose-700 dark:text-rose-300 border-rose-500/30"
          }`}>
            {t(`status.${String(user.status).toLowerCase()}`) || user.status}
          </span>
        </div>

        {/* Content Form */}
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {isCurrentlySuspended ? (
            <div className="p-3.5 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-800 dark:text-emerald-300 text-xs">
              <p className="font-medium leading-relaxed">
                {t("modals.suspend.reactivateBody")}
              </p>
            </div>
          ) : (
            <>
              <div className="space-y-2">
                <label className="text-xs font-semibold text-foreground uppercase tracking-wider block">
                  {t("modals.suspend.duration")}
                </label>
                <div className="grid grid-cols-2 gap-2">
                  {[
                    { id: "24h", label: t("modals.suspend.duration24h") },
                    { id: "7d", label: t("modals.suspend.duration7d") },
                    { id: "30d", label: t("modals.suspend.duration30d") },
                    { id: "indefinite", label: t("modals.suspend.durationIndefinite") },
                  ].map((opt) => (
                    <button
                      key={opt.id}
                      type="button"
                      onClick={() => setDuration(opt.id as any)}
                      className={`py-2 px-3 rounded-xl border text-xs font-semibold transition-all ${
                        duration === opt.id
                          ? "bg-primary text-primary-foreground border-primary shadow-xs"
                          : "border-border bg-card text-foreground hover:bg-secondary"
                      }`}
                    >
                      {opt.label}
                    </button>
                  ))}
                </div>
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-foreground uppercase tracking-wider block">
                  {t("modals.suspend.reason")}
                </label>
                <textarea
                  rows={3}
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  placeholder={t("modals.suspend.reasonPlaceholder")}
                  className="w-full p-3 rounded-xl border border-border bg-secondary/20 text-xs text-foreground placeholder:text-muted-foreground focus:ring-2 focus:ring-primary focus:outline-hidden resize-none"
                />
              </div>

              <div className="p-3 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-700 dark:text-rose-400 flex items-start gap-2.5 text-xs">
                <AlertOctagon className="size-4 shrink-0 mt-0.5" />
                <p className="text-[11px] leading-relaxed">
                  {t("modals.suspend.warning")}
                </p>
              </div>
            </>
          )}

          {/* Action Buttons */}
          <div className="pt-2 flex items-center justify-end gap-3">
            <button
              type="button"
              onClick={onClose}
              disabled={submitting}
              className="px-4 py-2.5 rounded-xl border border-border text-xs font-semibold text-foreground hover:bg-secondary transition-colors"
            >
              {t("actions.cancel")}
            </button>
            <button
              type="submit"
              disabled={submitting}
              className={`px-5 py-2.5 rounded-xl text-xs font-bold transition-colors flex items-center gap-2 shadow-xs ${
                isCurrentlySuspended
                  ? "bg-emerald-600 text-white hover:bg-emerald-700"
                  : "bg-destructive text-destructive-foreground hover:bg-destructive/90"
              }`}
            >
              {submitting ? (
                <>
                  <Loader2 className="size-3.5 animate-spin" />
                  <span>{t("modals.processing")}</span>
                </>
              ) : (
                <span>{isCurrentlySuspended ? t("modals.suspend.confirmReactivate") : t("modals.suspend.confirmSuspend")}</span>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
