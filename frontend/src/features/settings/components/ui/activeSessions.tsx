"use client";

import { useEffect, useState } from "react";
import { userService } from "@/services/user";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { useAuthContext } from "@/context/AuthContext";
import { useRouter } from "@/i18n/navigation";
import { Laptop, Smartphone, Globe, Shield, Trash2, RefreshCw, LogOut } from "lucide-react";

interface Session {
  id: string;
  userAgent?: string;
  ipAddress?: string;
  createdAt?: string;
  lastSeenAt?: string;
  isCurrent?: boolean;
}

function parseDevice(userAgent?: string) {
  if (!userAgent) return { label: "Unknown Device", icon: Globe };
  const ua = userAgent.toLowerCase();
  if (ua.includes("mobile") || ua.includes("android") || ua.includes("iphone")) {
    return { label: userAgent, icon: Smartphone };
  }
  return { label: userAgent, icon: Laptop };
}

export default function ActiveSessions() {
  const t = useTranslations("settings.sessions");
  const { logout } = useAuthContext();
  const router = useRouter();
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [revokingId, setRevokingId] = useState<string | null>(null);
  const [loggingOutAll, setLoggingOutAll] = useState(false);

  const fetchSessions = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await userService.getSessions();
      const raw: Session[] = Array.isArray(res) ? res : res?.data ?? [];
      setSessions(raw);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : t("error"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSessions();
  }, []);

  const handleDelete = async (id: string) => {
    setRevokingId(id);
    try {
      await userService.deleteSession(id);
      setSessions((prev) => prev.filter((s) => s.id !== id));
      toast.success(t("removedSuccess"));
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : t("error"));
    } finally {
      setRevokingId(null);
    }
  };

  const handleLogoutAll = async () => {
    setLoggingOutAll(true);
    try {
      await logout();
      router.push("/auth");
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : "Failed to sign out");
      setLoggingOutAll(false);
    }
  };

  return (
    <div className="flex flex-col gap-6 w-full">
      {/* ── Active Device Sessions List ── */}
      <div className="flex flex-col gap-3">
        {/* Header card with refresh action */}
        <div className="flex items-center justify-between pb-1">
          <p className="text-xs text-muted-foreground">{t("subtitle")}</p>
          <button
            type="button"
            onClick={fetchSessions}
            disabled={loading}
            className="inline-flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground transition-colors disabled:opacity-50 cursor-pointer"
          >
            <RefreshCw className={`size-3.5 ${loading ? "animate-spin" : ""}`} />
          </button>
        </div>

        {loading ? (
          <div className="flex flex-col gap-2.5">
            <div className="animate-pulse bg-secondary/50 rounded-xl h-16" />
            <div className="animate-pulse bg-secondary/50 rounded-xl h-16" />
          </div>
        ) : error ? (
          <div className="rounded-xl border border-destructive/20 bg-destructive/10 p-4 text-xs text-destructive">
            {error}
          </div>
        ) : sessions.length === 0 ? (
          <div className="rounded-xl border border-dashed border-border/80 bg-secondary/20 p-8 text-center text-xs text-muted-foreground">
            {t("empty")}
          </div>
        ) : (
          <div className="flex flex-col gap-2.5">
            {sessions.map((session) => {
              const device = parseDevice(session.userAgent);
              const DeviceIcon = device.icon;

              return (
                <div
                  key={session.id}
                  className={`bg-card border rounded-xl p-3.5 sm:p-4 flex items-center justify-between gap-3 transition-colors ${
                    session.isCurrent ? "border-emerald-500/30 bg-emerald-500/[0.02]" : "border-border/80"
                  }`}
                >
                  <div className="flex items-center gap-3 min-w-0">
                    <div
                      className={`size-9 rounded-lg flex items-center justify-center shrink-0 ${
                        session.isCurrent
                          ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
                          : "bg-secondary text-muted-foreground"
                      }`}
                    >
                      <DeviceIcon className="size-4.5" />
                    </div>

                    <div className="flex flex-col min-w-0 gap-0.5">
                      <div className="flex items-center gap-2">
                        <span className="text-xs sm:text-sm text-foreground font-medium truncate max-w-[240px] sm:max-w-md">
                          {session.userAgent ? session.userAgent : t("unknownDevice")}
                        </span>
                        {session.isCurrent && (
                          <span className="inline-flex items-center gap-1 px-1.5 py-0.2 rounded text-[10px] font-semibold bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20 shrink-0">
                            <Shield className="size-2.5" />
                            {t("current")}
                          </span>
                        )}
                      </div>

                      <div className="flex items-center gap-2 text-[11px] text-muted-foreground font-mono">
                        {session.ipAddress && <span>{session.ipAddress}</span>}
                        {session.lastSeenAt && (
                          <>
                            <span>•</span>
                            <span>{new Date(session.lastSeenAt).toLocaleDateString()}</span>
                          </>
                        )}
                      </div>
                    </div>
                  </div>

                  {!session.isCurrent && (
                    <button
                      type="button"
                      onClick={() => handleDelete(session.id)}
                      disabled={revokingId === session.id}
                      className="inline-flex items-center gap-1 px-2.5 py-1.5 rounded-lg text-xs font-medium text-destructive hover:bg-destructive/10 transition-colors disabled:opacity-40 disabled:cursor-not-allowed cursor-pointer shrink-0"
                      title={t("remove")}
                    >
                      <Trash2 className="size-3.5" />
                      <span className="hidden sm:inline">{t("remove")}</span>
                    </button>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* ── Sign Out Everywhere (Global Invalidation) ── */}
      <div className="pt-4 border-t border-border/70 flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-secondary/15 rounded-xl p-4 border border-border/60">
        <div className="min-w-0">
          <p className="text-xs sm:text-sm font-semibold text-foreground">{t("signOutAll")}</p>
          <p className="text-[11px] sm:text-xs text-muted-foreground mt-0.5 max-w-md leading-relaxed">
            {t("signOutAllDesc")}
          </p>
        </div>
        <button
          type="button"
          onClick={handleLogoutAll}
          disabled={loggingOutAll}
          className="inline-flex items-center justify-center gap-1.5 px-3.5 py-2 rounded-xl text-xs font-semibold border border-destructive/40 text-destructive hover:bg-destructive/10 transition-colors shadow-2xs shrink-0 cursor-pointer disabled:opacity-50"
        >
          <LogOut className="size-3.5" />
          <span>{t("signOutAllButton")}</span>
        </button>
      </div>
    </div>
  );
}
