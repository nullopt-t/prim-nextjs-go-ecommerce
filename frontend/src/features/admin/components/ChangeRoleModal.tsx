"use client";

import React, { useState } from "react";
import { 
  X, 
  ShieldAlert, 
  UserCheck, 
  Shield, 
  AlertTriangle,
  Loader2
} from "lucide-react";
import { AdminUser } from "@/services/admin";
import { ROLE_RANKS } from "@/lib/roles";
import { useTranslations } from "next-intl";

interface ChangeRoleModalProps {
  user: AdminUser;
  callerRole?: string;
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (userId: string, newRole: AdminUser["role"]) => Promise<void>;
}

const ROLES: { value: AdminUser["role"]; labelKey: string; descKey: string; color: string }[] = [
  { 
    value: "customer", 
    labelKey: "modals.changeRole.roleOptions.customer.label", 
    descKey: "modals.changeRole.roleOptions.customer.desc",
    color: "border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-900/40"
  },
  { 
    value: "vendor", 
    labelKey: "modals.changeRole.roleOptions.vendor.label", 
    descKey: "modals.changeRole.roleOptions.vendor.desc",
    color: "border-amber-300 dark:border-amber-700 bg-amber-50 dark:bg-amber-950/20"
  },
  { 
    value: "support", 
    labelKey: "modals.changeRole.roleOptions.support.label", 
    descKey: "modals.changeRole.roleOptions.support.desc",
    color: "border-teal-300 dark:border-teal-700 bg-teal-50 dark:bg-teal-950/20"
  },
  { 
    value: "admin", 
    labelKey: "modals.changeRole.roleOptions.admin.label", 
    descKey: "modals.changeRole.roleOptions.admin.desc",
    color: "border-blue-300 dark:border-blue-700 bg-blue-50 dark:bg-blue-950/20"
  },
];

export function ChangeRoleModal({ user, callerRole = "superadmin", isOpen, onClose, onConfirm }: ChangeRoleModalProps) {
  const t = useTranslations("admin");
  const callerRank = ROLE_RANKS[callerRole?.toLowerCase() || "customer"] || 1;
  const availableRoles = ROLES.filter((r) => (ROLE_RANKS[r.value] || 1) < callerRank);

  const [selectedRole, setSelectedRole] = useState<AdminUser["role"]>(
    availableRoles.some((r) => r.value === user.role) ? user.role : (availableRoles[0]?.value || "customer")
  );
  const [submitting, setSubmitting] = useState(false);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (selectedRole === user.role) {
      onClose();
      return;
    }
    setSubmitting(true);
    try {
      await onConfirm(user.id, selectedRole);
      onClose();
    } finally {
      setSubmitting(false);
    }
  };

  const isEscalation = selectedRole === "admin" && user.role !== "admin";

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs animate-in fade-in duration-150">
      <div 
        className="bg-card border border-border w-full max-w-lg rounded-2xl shadow-2xl overflow-hidden flex flex-col max-h-[90vh]"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="px-6 py-5 border-b border-border flex items-center justify-between bg-muted/30">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-purple-500/10 text-purple-600 dark:text-purple-400 border border-purple-500/20">
              <Shield className="size-5" />
            </div>
            <div>
              <h2 className="text-base font-bold text-foreground">{t("modals.changeRole.title")}</h2>
              <p className="text-xs text-muted-foreground">{t("modals.changeRole.subtitle")}</p>
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
          <span className="px-2.5 py-1 rounded-md text-[11px] font-bold uppercase tracking-wider bg-secondary text-secondary-foreground border border-border">
            {t("modals.changeRole.current", { role: user.role })}
          </span>
        </div>

        {/* Role Options Form */}
        <form onSubmit={handleSubmit} className="p-6 overflow-y-auto space-y-4 flex-1">
          <div className="space-y-2.5">
            <label className="text-xs font-semibold text-foreground uppercase tracking-wider block">
              {t("modals.changeRole.selectTier")}
            </label>
            <div className="space-y-2">
              {availableRoles.map((role) => (
                <label
                  key={role.value}
                  className={`flex items-start gap-3 p-3.5 rounded-xl border cursor-pointer transition-all ${
                    selectedRole === role.value
                      ? "ring-2 ring-primary border-primary " + role.color
                      : "border-border hover:bg-secondary/40"
                  }`}
                >
                  <input
                    type="radio"
                    name="userRole"
                    value={role.value}
                    checked={selectedRole === role.value}
                    onChange={() => setSelectedRole(role.value)}
                    className="mt-0.5 text-primary focus:ring-primary"
                  />
                  <div className="flex-1">
                    <div className="flex items-center justify-between">
                      <span className="font-bold text-xs text-foreground">{t(role.labelKey)}</span>
                      {role.value === user.role && (
                        <span className="text-[10px] text-muted-foreground font-medium">{t("modals.changeRole.currentLabel")}</span>
                      )}
                    </div>
                    <p className="text-[11px] text-muted-foreground mt-0.5 leading-relaxed">{t(role.descKey)}</p>
                  </div>
                </label>
              ))}
            </div>
          </div>

          {isEscalation && (
            <div className="p-3.5 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-700 dark:text-amber-400 flex items-start gap-2.5 text-xs">
              <AlertTriangle className="size-4 shrink-0 mt-0.5" />
              <div>
                <span className="font-bold">{t("modals.changeRole.escalationTitle")}</span>
                <p className="text-[11px] mt-0.5 text-amber-700/90 dark:text-amber-400/90">
                  {t("modals.changeRole.escalationBody")}
                </p>
              </div>
            </div>
          )}

          {/* Actions */}
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
              disabled={submitting || selectedRole === user.role}
              className="px-5 py-2.5 rounded-xl bg-primary text-primary-foreground text-xs font-bold hover:bg-primary/90 transition-colors flex items-center gap-2 shadow-xs disabled:opacity-50"
            >
              {submitting ? (
                <>
                  <Loader2 className="size-3.5 animate-spin" />
                  <span>{t("modals.updating")}</span>
                </>
              ) : (
                <span>{t("modals.changeRole.confirm")}</span>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
