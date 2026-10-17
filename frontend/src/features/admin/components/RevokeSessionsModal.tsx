"use client";

import React, { useState } from "react";
import { 
  X, 
  LogOut, 
  AlertTriangle,
  Loader2 
} from "lucide-react";
import { AdminUser } from "@/services/admin";
import { useTranslations } from "next-intl";

interface RevokeSessionsModalProps {
  user: AdminUser;
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (userId: string) => Promise<void>;
}

export function RevokeSessionsModal({ user, isOpen, onClose, onConfirm }: RevokeSessionsModalProps) {
  const t = useTranslations("admin");
  const [submitting, setSubmitting] = useState(false);

  if (!isOpen) return null;

  const handleRevoke = async () => {
    setSubmitting(true);
    try {
      await onConfirm(user.id);
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
        <div className="px-6 py-5 border-b border-border flex items-center justify-between bg-amber-500/10">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-amber-500/20 text-amber-600 dark:text-amber-400 border border-amber-500/30">
              <LogOut className="size-5" />
            </div>
            <div>
              <h2 className="text-base font-bold text-foreground">{t("modals.revoke.title")}</h2>
              <p className="text-xs text-muted-foreground">{t("modals.revoke.subtitle")}</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors"
          >
            <X className="size-4" />
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-4">
          <div className="p-3.5 rounded-xl bg-secondary/30 border border-border text-xs flex flex-col gap-1">
            <span className="font-semibold text-foreground">{user.fullName || t("modals.revoke.user")}</span>
            <span className="font-mono text-muted-foreground">{user.email}</span>
            <span className="font-mono text-[10px] text-muted-foreground mt-1">{t("modals.revoke.uuid", { id: user.id })}</span>
          </div>

          <div className="p-3.5 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-800 dark:text-amber-300 text-xs flex items-start gap-2.5">
            <AlertTriangle className="size-4 shrink-0 mt-0.5 text-amber-600 dark:text-amber-400" />
            <p className="text-[11px] leading-relaxed">
              {t("modals.revoke.body")}
            </p>
          </div>

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
              type="button"
              onClick={handleRevoke}
              disabled={submitting}
              className="px-5 py-2.5 rounded-xl bg-amber-600 text-white text-xs font-bold hover:bg-amber-700 transition-colors flex items-center gap-2 shadow-xs"
            >
              {submitting ? (
                <>
                  <Loader2 className="size-3.5 animate-spin" />
                  <span>{t("modals.revoking")}</span>
                </>
              ) : (
                <span>{t("modals.revoke.confirm")}</span>
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
