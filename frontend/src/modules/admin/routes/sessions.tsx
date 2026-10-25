"use client";

import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { useState, useEffect, useCallback } from "react";
import { MonitorSmartphone, RefreshCcw, ShieldAlert, X } from "lucide-react";
import { adminService, type StaffSession } from "@/modules/admin/api/adminService";
import { useLocale, useTranslations } from "next-intl";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { formatDateTime } from "@/shared/lib/datetime";

const STAFF_ROLES = ["all", "superadmin", "admin", "support", "vendor"] as const;

export function AdminSessions() {
  const t = useTranslations("admin");
  const locale = useLocale();
  const [sessions, setSessions] = useState<StaffSession[]>([]);
  const [roleFilter, setRoleFilter] = useState<string>("all");
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);
  const [revoking, setRevoking] = useState<string | null>(null);
  const [detail, setDetail] = useState<StaffSession | null>(null);

  const load = useCallback(async () => {
    try {
      setLoading(true);
      setLoadError(false);
      const res = await adminService.getStaffSessions(
        roleFilter === "all" ? undefined : roleFilter
      );
      setSessions(Array.isArray(res?.data) ? res.data : []);
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [roleFilter]);

  useEffect(() => {
    load();
  }, [load]);

  const handleRevoke = async (session: StaffSession) => {
    if (!confirm(t("sessions.confirmRevoke", { email: session.email }))) return;
    try {
      setRevoking(session.id);
      await adminService.revokeStaffSession(session.id);
      notify.success(t("sessions.toast.revoked"));
      setDetail(null);
      await load();
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("sessions.toast.revokeFailed")));
    } finally {
      setRevoking(null);
    }
  };

  const roleBadge = (role: string) =>
    role === "superadmin"
      ? "bg-rose-500/10 text-rose-600 border-rose-500/20"
      : role === "admin"
        ? "bg-blue-500/10 text-blue-600 border-blue-500/20"
        : role === "support"
          ? "bg-teal-500/10 text-teal-600 border-teal-500/20"
          : "bg-amber-500/10 text-amber-600 border-amber-500/20";

  if (loadError) {
    return <ErrorState onRetry={load} />;
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t("sessions.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">{t("sessions.subtitle")}</p>
        </div>
        <button
          onClick={load}
          disabled={loading}
          className="flex items-center gap-2 px-4 py-2.5 border border-border bg-secondary/30 hover:bg-secondary text-foreground text-xs font-semibold rounded-xl transition-colors disabled:opacity-50 self-start md:self-auto"
        >
          <RefreshCcw className={`size-4 ${loading ? "animate-spin" : ""}`} />
          <span>{t("sessions.refresh")}</span>
        </button>
      </div>

      <div className="bg-background border border-border rounded-2xl p-4 shadow-xs">
        <div className="flex flex-wrap items-center gap-2">
          {STAFF_ROLES.map((role) => (
            <button
              key={role}
              onClick={() => setRoleFilter(role)}
              className={`px-3 py-1.5 rounded-lg text-xs font-semibold border transition-colors ${
                roleFilter === role
                  ? "bg-accent-brand text-white border-accent-brand"
                  : "border-border text-muted-foreground hover:text-foreground hover:bg-secondary/40"
              }`}
            >
              {role === "all" ? t("sessions.allRoles") : t(`roles.${role}`)}
            </button>
          ))}
        </div>
      </div>

      <div className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs">
        <div className="overflow-x-auto">
          <ResponsiveTable className="w-full sm:min-w-max text-start text-xs">
            <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
              <tr>
                <th className="py-3.5 ps-4 text-start">{t("sessions.colUser")}</th>
                <th className="py-3.5 text-start">{t("sessions.colRole")}</th>
                <th className="py-3.5 text-start">{t("sessions.colDevice")}</th>
                <th className="py-3.5 text-start">{t("sessions.colIp")}</th>
                <th className="py-3.5 text-start">{t("sessions.colCreated")}</th>
                <th className="py-3.5 text-start">{t("sessions.colLastActive")}</th>
                <th className="py-3.5 pe-4 text-end">{t("sessions.colAction")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {sessions.map((session) => (
                <tr
                  key={session.id}
                  onClick={() => setDetail(session)}
                  className="hover:bg-secondary/20 transition-colors cursor-pointer"
                >
                  <td className="py-3.5 ps-4">
                    <div className="flex items-center gap-2">
                      <span className="font-semibold text-foreground truncate max-w-[180px]">
                        {session.fullName || session.email}
                      </span>
                      {session.current && (
                        <span className="px-1.5 py-0.5 rounded text-[9px] font-bold bg-emerald-500/10 text-emerald-600 border border-emerald-500/20">
                          {t("sessions.thisDevice")}
                        </span>
                      )}
                    </div>
                    <div className="font-mono text-[10px] text-muted-foreground truncate max-w-[220px]">
                      {session.email}
                    </div>
                  </td>
                  <td className="py-3.5">
                    <span className={`px-2 py-0.5 rounded-md text-[10px] font-bold border ${roleBadge(session.role)}`}>
                      {t(`roles.${session.role}`)}
                    </span>
                  </td>
                  <td className="py-3.5 text-muted-foreground">
                    <span className="block truncate max-w-[220px]">{session.userAgent || "—"}</span>
                  </td>
                  <td className="py-3.5 font-mono text-[10px] text-muted-foreground">
                    {session.ipAddress || "—"}
                  </td>
                  <td className="py-3.5 text-muted-foreground whitespace-nowrap">
                    {formatDateTime(session.createdAt, locale)}
                  </td>
                  <td className="py-3.5 text-muted-foreground whitespace-nowrap">
                    {formatDateTime(session.lastActive, locale)}
                  </td>
                  <td className="py-3.5 pe-4 text-end" onClick={(e) => e.stopPropagation()}>
                    <button
                      onClick={() => handleRevoke(session)}
                      disabled={session.current || revoking === session.id}
                      title={session.current ? t("sessions.cannotRevokeCurrent") : t("sessions.revoke")}
                      className="px-2.5 py-1 text-[11px] font-semibold text-destructive bg-destructive/10 hover:bg-destructive/20 rounded-lg transition-colors inline-flex items-center gap-1 disabled:opacity-40 disabled:cursor-not-allowed"
                    >
                      <X className="size-3" />
                      <span>{t("sessions.revoke")}</span>
                    </button>
                  </td>
                </tr>
              ))}
              {sessions.length === 0 && !loading && (
                <tr>
                  <td colSpan={7} className="py-10 text-center text-muted-foreground">
                    <MonitorSmartphone className="size-6 mx-auto mb-2 opacity-50" />
                    {t("sessions.empty")}
                  </td>
                </tr>
              )}
            </tbody>
          </ResponsiveTable>
        </div>
        {loading && sessions.length === 0 && (
          <div className="py-10 text-center text-xs text-muted-foreground">
            {t("sessions.loading")}
          </div>
        )}
      </div>

      {detail && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-background border border-border rounded-2xl max-w-lg w-full max-h-[85vh] overflow-y-auto shadow-2xl">
            <div className="flex items-start justify-between gap-4 px-5 py-4 border-b border-border">
              <div className="min-w-0">
                <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
                  {t("sessions.detailTitle")}
                </div>
                <h3 className="font-bold text-base text-foreground flex items-center gap-2">
                  <ShieldAlert className="size-4 text-accent-brand shrink-0" />
                  <span className="truncate">{detail.fullName || detail.email}</span>
                </h3>
              </div>
              <button
                onClick={() => setDetail(null)}
                className="-me-1 shrink-0 rounded-lg p-1.5 text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors"
                aria-label={t("actions.close")}
              >
                <X className="size-5" />
              </button>
            </div>

            <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-4 px-5 py-5 text-xs">
              <div className="sm:col-span-2 space-y-0.5">
                <dt className="font-medium text-muted-foreground">{t("sessions.colUser")}</dt>
                <dd className="text-foreground" dir="ltr">
                  {detail.fullName ? `${detail.fullName} (${detail.email})` : detail.email}
                </dd>
              </div>
              <div className="space-y-0.5">
                <dt className="font-medium text-muted-foreground">{t("sessions.colRole")}</dt>
                <dd>
                  <span
                    className={`inline-block px-2 py-0.5 rounded-md text-[10px] font-bold border ${roleBadge(detail.role)}`}
                  >
                    {t(`roles.${detail.role}`)}
                  </span>
                </dd>
              </div>
              <div className="space-y-0.5">
                <dt className="font-medium text-muted-foreground">{t("sessions.colIp")}</dt>
                <dd className="text-foreground font-mono" dir="ltr">
                  {detail.ipAddress || "—"}
                </dd>
              </div>
              <div className="space-y-0.5">
                <dt className="font-medium text-muted-foreground">{t("sessions.colCreated")}</dt>
                <dd className="text-foreground">{formatDateTime(detail.createdAt, locale)}</dd>
              </div>
              <div className="space-y-0.5">
                <dt className="font-medium text-muted-foreground">{t("sessions.colLastActive")}</dt>
                <dd className="text-foreground">{formatDateTime(detail.lastActive, locale)}</dd>
              </div>
              <div className="sm:col-span-2 space-y-0.5">
                <dt className="font-medium text-muted-foreground">{t("sessions.sessionId")}</dt>
                <dd className="text-foreground font-mono break-all" dir="ltr">
                  {detail.id}
                </dd>
              </div>
              <div className="sm:col-span-2 space-y-0.5">
                <dt className="font-medium text-muted-foreground">{t("sessions.colDevice")}</dt>
                <dd className="text-foreground break-all" dir="ltr">
                  {detail.userAgent || "—"}
                </dd>
              </div>
            </dl>

            <div className="flex justify-end px-5 pb-5">
              <button
                onClick={() => handleRevoke(detail)}
                disabled={detail.current || revoking === detail.id}
                title={detail.current ? t("sessions.cannotRevokeCurrent") : undefined}
                className="px-4 py-2 text-xs font-semibold text-white bg-destructive rounded-xl hover:bg-destructive/90 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
              >
                {t("sessions.revoke")}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
