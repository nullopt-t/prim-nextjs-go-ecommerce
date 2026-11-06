"use client";

import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { useState, useEffect, useCallback } from "react";
import { X } from "lucide-react";
import { adminService, type StaffSession } from "@/modules/admin/api/adminService";
import { useLocale, useTranslations } from "next-intl";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
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

  const roleLabel = (role: string) => t(`roles.${role}`);

  if (loadError) {
    return <ErrorState onRetry={load} />;
  }

  return (
    <div className="space-y-5">
      <PageHeader
        title={t("sessions.title")}
        actions={
          <button
            onClick={load}
            disabled={loading}
            className="rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground transition-colors hover:bg-secondary disabled:opacity-50"
          >
            {t("sessions.refresh")}
          </button>
        }
      />

      <div className="flex items-center justify-between gap-3">
        <select
          value={roleFilter}
          onChange={(e) => setRoleFilter(e.target.value)}
          className="rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
          aria-label={t("sessions.colRole")}
        >
          {STAFF_ROLES.map((role) => (
            <option key={role} value={role}>
              {role === "all" ? t("sessions.allRoles") : roleLabel(role)}
            </option>
          ))}
        </select>
        <span className="text-xs text-muted-foreground">
          {t("sessions.count", { count: sessions.length })}
        </span>
      </div>

      <div className="overflow-hidden rounded-lg border border-border bg-background">
        <div className="overflow-x-auto">
          <ResponsiveTable className="w-full text-start text-sm md:min-w-max">
            <thead className="text-xs text-muted-foreground">
              <tr className="border-b border-border">
                <th className="py-2.5 ps-4 text-start font-medium">{t("sessions.colUser")}</th>
                <th className="py-2.5 text-start font-medium">{t("sessions.colRole")}</th>
                <th className="py-2.5 text-start font-medium">{t("sessions.colDevice")}</th>
                <th className="py-2.5 text-start font-medium">{t("sessions.colIp")}</th>
                <th className="py-2.5 text-start font-medium">{t("sessions.colCreated")}</th>
                <th className="py-2.5 text-start font-medium">{t("sessions.colLastActive")}</th>
                <th className="py-2.5 pe-4 text-end font-medium">{t("sessions.colAction")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {sessions.map((session) => (
                <tr
                  key={session.id}
                  onClick={() => setDetail(session)}
                  className="cursor-pointer transition-colors hover:bg-secondary/30"
                >
                  <td className="py-3 ps-4">
                    <div className="flex items-center gap-2">
                      <span className="max-w-[180px] truncate font-medium text-foreground">
                        {session.fullName || session.email}
                      </span>
                      {session.current ? <StatusPill tone="success">{t("sessions.thisDevice")}</StatusPill> : null}
                    </div>
                    <div className="max-w-[220px] truncate font-mono text-xs text-muted-foreground">
                      {session.email}
                    </div>
                  </td>
                  <td className="py-3">
                    <StatusPill tone={statusTone(session.role)}>{roleLabel(session.role)}</StatusPill>
                  </td>
                  <td className="py-3 text-xs text-muted-foreground">
                    <span className="block max-w-[220px] truncate">{session.userAgent || "—"}</span>
                  </td>
                  <td className="py-3 font-mono text-xs text-muted-foreground">
                    {session.ipAddress || "—"}
                  </td>
                  <td className="whitespace-nowrap py-3 text-xs text-muted-foreground">
                    {formatDateTime(session.createdAt, locale) || "—"}
                  </td>
                  <td className="whitespace-nowrap py-3 text-xs text-muted-foreground">
                    {formatDateTime(session.lastActive, locale) || "—"}
                  </td>
                  <td className="py-3 pe-4 text-end" onClick={(e) => e.stopPropagation()}>
                    <button
                      onClick={() => handleRevoke(session)}
                      disabled={session.current || revoking === session.id}
                      title={session.current ? t("sessions.cannotRevokeCurrent") : t("sessions.revoke")}
                      className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive disabled:cursor-not-allowed disabled:opacity-40"
                    >
                      {t("sessions.revoke")}
                    </button>
                  </td>
                </tr>
              ))}
              {sessions.length === 0 && !loading ? (
                <tr>
                  <td colSpan={7} className="py-10 text-center text-sm text-muted-foreground">
                    {t("sessions.empty")}
                  </td>
                </tr>
              ) : null}
            </tbody>
          </ResponsiveTable>
        </div>
        {loading && sessions.length === 0 ? (
          <div className="py-10 text-center text-sm text-muted-foreground">
            {t("sessions.loading")}
          </div>
        ) : null}
      </div>

      {detail ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="max-h-[85vh] w-full max-w-lg space-y-4 overflow-y-auto rounded-lg border border-border bg-background p-5 shadow-xl">
            <div className="flex items-start justify-between gap-4 border-b border-border pb-3">
              <div className="min-w-0">
                <div className="text-xs font-medium text-muted-foreground">{t("sessions.detailTitle")}</div>
                <h3 className="mt-0.5 truncate text-sm font-semibold text-foreground">
                  {detail.fullName || detail.email}
                </h3>
              </div>
              <button
                onClick={() => setDetail(null)}
                className="shrink-0 text-muted-foreground hover:text-foreground"
                aria-label={t("actions.close")}
              >
                <X className="size-4" />
              </button>
            </div>

            <dl className="grid grid-cols-1 gap-x-6 gap-y-4 text-sm sm:grid-cols-2">
              <div className="space-y-0.5 sm:col-span-2">
                <dt className="text-xs text-muted-foreground">{t("sessions.colUser")}</dt>
                <dd className="text-foreground" dir="ltr">
                  {detail.fullName ? `${detail.fullName} (${detail.email})` : detail.email}
                </dd>
              </div>
              <div className="space-y-0.5">
                <dt className="text-xs text-muted-foreground">{t("sessions.colRole")}</dt>
                <dd>
                  <StatusPill tone={statusTone(detail.role)}>{roleLabel(detail.role)}</StatusPill>
                </dd>
              </div>
              <div className="space-y-0.5">
                <dt className="text-xs text-muted-foreground">{t("sessions.colIp")}</dt>
                <dd className="font-mono text-foreground" dir="ltr">
                  {detail.ipAddress || "—"}
                </dd>
              </div>
              <div className="space-y-0.5">
                <dt className="text-xs text-muted-foreground">{t("sessions.colCreated")}</dt>
                <dd className="text-foreground">{formatDateTime(detail.createdAt, locale) || "—"}</dd>
              </div>
              <div className="space-y-0.5">
                <dt className="text-xs text-muted-foreground">{t("sessions.colLastActive")}</dt>
                <dd className="text-foreground">{formatDateTime(detail.lastActive, locale) || "—"}</dd>
              </div>
              <div className="space-y-0.5 sm:col-span-2">
                <dt className="text-xs text-muted-foreground">{t("sessions.sessionId")}</dt>
                <dd className="break-all font-mono text-xs text-foreground" dir="ltr">
                  {detail.id}
                </dd>
              </div>
              <div className="space-y-0.5 sm:col-span-2">
                <dt className="text-xs text-muted-foreground">{t("sessions.colDevice")}</dt>
                <dd className="break-all text-foreground" dir="ltr">
                  {detail.userAgent || "—"}
                </dd>
              </div>
            </dl>

            <div className="flex justify-end border-t border-border pt-3">
              <button
                onClick={() => handleRevoke(detail)}
                disabled={detail.current || revoking === detail.id}
                title={detail.current ? t("sessions.cannotRevokeCurrent") : undefined}
                className="rounded-md bg-destructive px-3 py-2 text-sm font-medium text-white transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-40"
              >
                {t("sessions.revoke")}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
