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
import { toast } from "sonner";

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

const ROLE_RANKS: Record<string, number> = {
  customer: 1,
  vendor: 2,
  support: 3,
  admin: 4,
  superadmin: 5,
};

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

  if (!isOpen || !user) return null;

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    toast.success("User ID copied to clipboard");
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
                    {user.fullName || "Unnamed Customer"}
                  </h2>
                  {user.isEmailVerified && (
                    <span title="Verified Email">
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
                <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Access Role</span>
                <span className="text-xs font-bold text-foreground capitalize flex items-center gap-1.5">
                  <Shield className="size-3.5 text-purple-500" />
                  {user.role}
                </span>
              </div>
              <div className="p-3.5 rounded-xl bg-secondary/30 border border-border flex flex-col gap-1">
                <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Account State</span>
                <span className={`text-xs font-bold capitalize flex items-center gap-1.5 ${
                  user.status === "active" ? "text-emerald-600 dark:text-emerald-400" : "text-rose-600 dark:text-rose-400"
                }`}>
                  <span className={`size-2 rounded-full ${user.status === "active" ? "bg-emerald-500" : "bg-rose-500"}`} />
                  {user.status}
                </span>
              </div>
            </div>

            {/* Account Identifiers */}
            <div className="space-y-3">
              <h3 className="text-xs font-bold text-foreground uppercase tracking-wider">Identification</h3>
              <div className="p-4 rounded-2xl bg-secondary/20 border border-border space-y-3 text-xs">
                <div>
                  <span className="text-muted-foreground block text-[11px] mb-1">Database UUID</span>
                  <div className="flex items-center justify-between bg-card p-2 rounded-lg border border-border">
                    <span className="font-mono text-[11px] text-foreground truncate">{user.id}</span>
                    <button
                      onClick={() => copyToClipboard(user.id)}
                      className="p-1 rounded text-muted-foreground hover:text-foreground transition-colors ml-2"
                      title="Copy ID"
                    >
                      {copied ? <Check className="size-3.5 text-emerald-500" /> : <Copy className="size-3.5" />}
                    </button>
                  </div>
                </div>

                <div className="flex justify-between items-center pt-2 border-t border-border/60">
                  <span className="text-muted-foreground">Email Verification</span>
                  <span className={`font-semibold flex items-center gap-1 text-[11px] ${
                    user.isEmailVerified ? "text-emerald-600 dark:text-emerald-400" : "text-amber-600 dark:text-amber-400"
                  }`}>
                    {user.isEmailVerified ? (
                      <>
                        <CheckCircle2 className="size-3.5" />
                        Verified
                      </>
                    ) : (
                      <>
                        <XCircle className="size-3.5" />
                        Unverified
                      </>
                    )}
                  </span>
                </div>
              </div>
            </div>

            {/* Timestamps & Security Details */}
            <div className="space-y-3">
              <h3 className="text-xs font-bold text-foreground uppercase tracking-wider">Security & Timestamps</h3>
              <div className="p-4 rounded-2xl bg-secondary/20 border border-border space-y-2.5 text-xs">
                <div className="flex justify-between items-center">
                  <span className="text-muted-foreground flex items-center gap-1.5">
                    <Calendar className="size-3.5" /> Registered
                  </span>
                  <span className="font-medium text-foreground">
                    {new Date(user.createdAt).toLocaleDateString(undefined, { 
                      month: "short", 
                      day: "numeric", 
                      year: "numeric" 
                    })}
                  </span>
                </div>

                <div className="flex justify-between items-center">
                  <span className="text-muted-foreground flex items-center gap-1.5">
                    <Clock className="size-3.5" /> Last Modified
                  </span>
                  <span className="font-medium text-foreground">
                    {new Date(user.updatedAt).toLocaleDateString(undefined, { 
                      month: "short", 
                      day: "numeric", 
                      year: "numeric" 
                    })}
                  </span>
                </div>

                {user.suspendedUntil && (
                  <div className="flex justify-between items-center pt-2 border-t border-border/60 text-rose-600 dark:text-rose-400">
                    <span className="font-semibold">Suspension Expires</span>
                    <span className="font-mono text-[11px]">
                      {new Date(user.suspendedUntil).toLocaleString()}
                    </span>
                  </div>
                )}

                {user.lockedUntil && (
                  <div className="flex justify-between items-center pt-2 border-t border-border/60 text-amber-600 dark:text-amber-400">
                    <span className="font-semibold">Security Lock Expires</span>
                    <span className="font-mono text-[11px]">
                      {new Date(user.lockedUntil).toLocaleTimeString()}
                    </span>
                  </div>
                )}
              </div>
            </div>

            {/* Quick Actions Panel */}
            <div className="space-y-3">
              <h3 className="text-xs font-bold text-foreground uppercase tracking-wider">Administrative Actions</h3>
              {isSelf ? (
                <div className="p-3.5 rounded-xl bg-secondary/30 border border-border text-center">
                  <p className="text-xs text-muted-foreground">
                    This is your account. Administrative role, suspension, and forced session actions cannot be applied to yourself.
                  </p>
                </div>
              ) : !canManage ? (
                <div className="p-3.5 rounded-xl bg-secondary/30 border border-border text-center">
                  <p className="text-xs text-muted-foreground">
                    This account possesses equal or higher administrative privileges ({user.role}). Actions cannot be applied to this account.
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
                      Change Permission Role
                    </span>
                    <span className="text-[10px] text-muted-foreground uppercase">Superadmin</span>
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
                      {isSuspended ? "Reactivate Account" : "Suspend Account"}
                    </span>
                    <span className="text-[10px] opacity-75 uppercase">Superadmin</span>
                  </button>

                  <button
                    onClick={() => onRevokeSessions(user)}
                    className="w-full py-2.5 px-4 rounded-xl border border-border bg-card hover:bg-secondary text-foreground text-xs font-semibold transition-colors flex items-center justify-between"
                  >
                    <span className="flex items-center gap-2">
                      <LogOut className="size-4 text-amber-500" />
                      Force Log Out (Revoke Sessions)
                    </span>
                    <span className="text-[10px] text-muted-foreground uppercase">Superadmin</span>
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
              Close Inspector
            </button>
          </div>

        </div>
      </div>
    </div>
  );
}
