"use client";

import React, { useState } from "react";
import { 
  X, 
  Shield, 
  Mail, 
  Calendar, 
  Clock, 
  CheckCircle2, 
  XCircle, 
  Copy, 
  Check, 
  UserCheck, 
  UserX, 
  LogOut, 
  ShoppingBag,
  ExternalLink 
} from "lucide-react";
import { AdminUser } from "@/services/admin";
import { useTranslations } from "next-intl";
import { notify } from "@/lib/notify";
import { formatDateTime } from "@/lib/datetime";
import { ROLE_RANKS } from "@/lib/roles";

interface UserSlideOverDrawerProps {
  user: AdminUser | null;
  currentUserId?: string;
  currentUserRole?: string;
  isOpen: boolean;
  onClose: () => void;
  onChangeRole: (user: AdminUser) => void;
  onSuspend: (user: AdminUser) => void;
  onRevokeSessions: (user: AdminUser) => void;
}

export function UserSlideOverDrawer({
  user,
  currentUserId,
  currentUserRole,
  isOpen,
  onClose,
  onChangeRole,
  onSuspend,
  onRevokeSessions,
}: UserSlideOverDrawerProps) {
  const [copied, setCopied] = useState(false);
  const t = useTranslations("admin");

  if (!isOpen || !user) return null;

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    notify.success(t("modals.drawer.toast.copied"));
    setTimeout(() => setCopied(false), 2000);
  };

  const isSuspended = user.status === "suspended";
  const isSelf = currentUserId && user.id && String(currentUserId) === String(user.id);
  const callerRank = ROLE_RANKS[currentUserRole?.toLowerCase() || "customer"] || 1;
  const targetRank = ROLE_RANKS[user.role?.toLowerCase() || "customer"] || 1;
  const canManage = !isSelf && callerRank > targetRank;

  return (
    <div className="fixed inset-0 z-40 overflow-hidden">
      {/* Backdrop */}
      <div 
        className="absolute inset-0 bg-black/40 backdrop-blur-xs transition-opacity animate-in fade-in"
        onClick={onClose}
      />

      <div className="fixed inset-y-0 right-0 max-w-full flex pl-10">
        <div className="w-screen max-w-md bg-card border-l border-border shadow-2xl flex flex-col justify-between animate-in slide-in-from-right duration-200">
          
          {/* Header */}
          <div className="p-6 border-b border-border bg-muted/20 flex items-start justify-between">
            <div className="flex items-center gap-3.5">
              <div className="size-12 rounded-2xl bg-primary/10 border border-primary/20 text-primary flex items-center justify-center font-bold text-base shadow-inner">
                {user.fullName ? user.fullName.slice(0, 2).toUpperCase() : user.email.slice(0, 2).toUpperCase()}
              </div>
              <div className="flex flex-col">
                <div className="flex items-center gap-2">
                  <h2 className="text-base font-bold text-foreground truncate max-w-[200px]">
                    {user.fullName || t("modals.drawer.unnamed")}
                  </h2>
                  {user.isEmailVerified && (
                    <span title={t("modals.drawer.verifiedEmail")}>
                      <CheckCircle2 className="size-4 text-emerald-500" />
                    </span>
                  )}
                </div>
                <span className="text-xs text-muted-foreground truncate max-w-[200px]">{user.email}</span>
              </div>
            </div>

            <button
              onClick={onClose}
              className="p-2 rounded-xl text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors"
            >
              <X className="size-4" />
            </button>
          </div>

          {/* Drawer Body */}
          <div className="flex-1 overflow-y-auto p-6 space-y-6">
            
            {/* Quick Status Bar */}
            <div className="grid grid-cols-2 gap-3">
              <div className="p-3.5 rounded-xl bg-secondary/30 border border-border flex flex-col gap-1">
                <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">{t("modals.drawer.accessRole")}</span>
                <span className="text-xs font-bold text-foreground capitalize">
                  {t(`roles.${user.role}`) || user.role}
                </span>
              </div>
              <div className="p-3.5 rounded-xl bg-secondary/30 border border-border flex flex-col gap-1">
                <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">{t("modals.drawer.accountState")}</span>
                <span className="text-xs font-bold capitalize text-foreground">
                  {t(`status.${String(user.status).toLowerCase()}`) || user.status}
                </span>
              </div>
            </div>

            {/* Account Identifiers */}
            <div className="space-y-3">
              <h3 className="text-xs font-bold text-foreground uppercase tracking-wider">{t("modals.drawer.identification")}</h3>
              <div className="p-4 rounded-2xl bg-secondary/20 border border-border space-y-3 text-xs">
                <div>
                  <span className="text-muted-foreground block text-[11px] mb-1">{t("modals.drawer.databaseUuid")}</span>
                  <div className="flex items-center justify-between bg-card p-2 rounded-lg border border-border">
                    <span className="font-mono text-[11px] text-foreground truncate">{user.id}</span>
                    <button
                      onClick={() => copyToClipboard(user.id)}
                      className="p-1 rounded text-muted-foreground hover:text-foreground transition-colors ms-2"
                      title={t("modals.drawer.copyId")}
                    >
                      {copied ? <Check className="size-3.5 text-emerald-500" /> : <Copy className="size-3.5" />}
                    </button>
                  </div>
                </div>

                <div className="flex justify-between items-center pt-2 border-t border-border/60">
                  <span className="text-muted-foreground">{t("modals.drawer.emailVerification")}</span>
                  <span className={`font-semibold flex items-center gap-1 text-[11px] ${
                    user.isEmailVerified ? "text-emerald-600 dark:text-emerald-400" : "text-amber-600 dark:text-amber-400"
                  }`}>
                    {user.isEmailVerified ? (
                      <>
                        <CheckCircle2 className="size-3.5" />
                        {t("status.verified")}
                      </>
                    ) : (
                      <>
                        <XCircle className="size-3.5" />
                        {t("status.unverified")}
                      </>
                    )}
                  </span>
                </div>
              </div>
            </div>

            {/* Timestamps & Security Details */}
            <div className="space-y-3">
              <h3 className="text-xs font-bold text-foreground uppercase tracking-wider">{t("modals.drawer.securityTimestamps")}</h3>
              <div className="p-4 rounded-2xl bg-secondary/20 border border-border space-y-2.5 text-xs">
                <div className="flex justify-between items-center">
                  <span className="text-muted-foreground flex items-center gap-1.5">
                    <Calendar className="size-3.5" /> {t("modals.drawer.registered")}
                  </span>
                  <span className="font-medium text-foreground">
                    {formatDateTime(user.createdAt)}
                  </span>
                </div>

                <div className="flex justify-between items-center">
                  <span className="text-muted-foreground flex items-center gap-1.5">
                    <Clock className="size-3.5" /> {t("modals.drawer.lastModified")}
                  </span>
                  <span className="font-medium text-foreground">
                    {formatDateTime(user.updatedAt)}
                  </span>
                </div>

                {user.suspendedUntil && (
                  <div className="flex justify-between items-center pt-2 border-t border-border/60 text-rose-600 dark:text-rose-400">
                    <span className="font-semibold">{t("modals.drawer.suspensionExpires")}</span>
                    <span className="font-mono text-[11px]">
                      {formatDateTime(user.suspendedUntil)}
                    </span>
                  </div>
                )}

                {user.lockedUntil && (
                  <div className="flex justify-between items-center pt-2 border-t border-border/60 text-amber-600 dark:text-amber-400">
                    <span className="font-semibold">{t("modals.drawer.lockExpires")}</span>
                    <span className="font-mono text-[11px]">
                      {formatDateTime(user.lockedUntil)}
                    </span>
                  </div>
                )}
              </div>
            </div>

            {/* Quick Actions Panel */}
            <div className="space-y-3">
              <h3 className="text-xs font-bold text-foreground uppercase tracking-wider">{t("modals.drawer.adminActions")}</h3>
              {isSelf ? (
                <div className="p-3.5 rounded-xl bg-secondary/30 border border-border text-center">
                  <p className="text-xs text-muted-foreground">
                    {t("modals.drawer.selfNote")}
                  </p>
                </div>
              ) : !canManage ? (
                <div className="p-3.5 rounded-xl bg-secondary/30 border border-border text-center">
                  <p className="text-xs text-muted-foreground">
                    {t("modals.drawer.higherPrivNote", { role: user.role })}
                  </p>
                </div>
              ) : (
                <div className="flex flex-col gap-2">
                  <button
                    onClick={() => onChangeRole(user)}
                    className="w-full py-2.5 px-4 rounded-xl border border-border bg-card hover:bg-secondary text-foreground text-xs font-semibold transition-colors flex items-center justify-between"
                  >
                    <span className="flex items-center gap-2">
                      <Shield className="size-4 text-purple-500" />
                      {t("modals.drawer.changeRole")}
                    </span>
                    <span className="text-[10px] text-muted-foreground uppercase">{t("roles.superadmin")}</span>
                  </button>

                  <button
                    onClick={() => onSuspend(user)}
                    className={`w-full py-2.5 px-4 rounded-xl border text-xs font-semibold transition-colors flex items-center justify-between ${
                      isSuspended
                        ? "border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 hover:bg-emerald-500/20"
                        : "border-rose-500/30 bg-rose-500/10 text-rose-700 dark:text-rose-300 hover:bg-rose-500/20"
                    }`}
                  >
                    <span className="flex items-center gap-2">
                      {isSuspended ? <UserCheck className="size-4" /> : <UserX className="size-4" />}
                      {isSuspended ? t("actions.reactivate") : t("actions.suspend")}
                    </span>
                    <span className="text-[10px] opacity-75 uppercase">{t("roles.superadmin")}</span>
                  </button>

                  <button
                    onClick={() => onRevokeSessions(user)}
                    className="w-full py-2.5 px-4 rounded-xl border border-border bg-card hover:bg-secondary text-foreground text-xs font-semibold transition-colors flex items-center justify-between"
                  >
                    <span className="flex items-center gap-2">
                      <LogOut className="size-4 text-amber-500" />
                      {t("modals.drawer.forceLogout")}
                    </span>
                    <span className="text-[10px] text-muted-foreground uppercase">{t("roles.superadmin")}</span>
                  </button>
                </div>
              )}
            </div>

          </div>

          {/* Footer */}
          <div className="p-4 border-t border-border bg-muted/20 flex justify-end">
            <button
              onClick={onClose}
              className="py-2 px-5 rounded-xl bg-secondary hover:bg-secondary/80 text-foreground text-xs font-bold transition-colors"
            >
              {t("modals.drawer.close")}
            </button>
          </div>

        </div>
      </div>
    </div>
  );
}
