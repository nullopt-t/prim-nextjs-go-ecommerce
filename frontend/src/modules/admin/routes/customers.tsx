"use client";

import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { useState, useEffect, useCallback } from "react";
import { Search } from "lucide-react";
import { adminService, AdminUser } from "@/modules/admin/api/adminService";
import { CursorPager, useCursorPagination } from "@/shared/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/shared/lib/pagination";
import { UserSlideOverDrawer } from "../components/UserSlideOverDrawer";
import { ChangeRoleModal } from "../components/ChangeRoleModal";
import { SuspendUserModal } from "../components/SuspendUserModal";
import { RevokeSessionsModal } from "../components/RevokeSessionsModal";
import { useAuthContext } from "@/modules/auth/model/authContext";
import { useLocale, useTranslations } from "next-intl";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { ROLE_RANKS } from "@/shared/lib/roles";
import { CustomSelect } from "@/shared/ui/customSelect";

export function AdminCustomers() {
  const { user: currentUser } = useAuthContext();
  const t = useTranslations("admin");
  const locale = useLocale();
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);

  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [roleFilter, setRoleFilter] = useState("all");
  const [statusFilter, setStatusFilter] = useState("all");
  const { cursor, hasPrev, next, prev, reset, page } = useCursorPagination();

  // Active User Drawer & Modals state
  const [inspectUser, setInspectUser] = useState<AdminUser | null>(null);
  const [roleModalUser, setRoleModalUser] = useState<AdminUser | null>(null);
  const [suspendModalUser, setSuspendModalUser] = useState<AdminUser | null>(null);
  const [revokeModalUser, setRevokeModalUser] = useState<AdminUser | null>(null);
  const [loadError, setLoadError] = useState(false);

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      reset(); // restart pagination on search
    }, 350);
    return () => clearTimeout(timer);
  }, [search, reset]);

  const fetchUsers = useCallback(async () => {
    setLoading(true);
    setLoadError(false);
    try {
      const res = await adminService.getUsers({
        cursor: cursor || undefined,
        pageSize: 10,
        search: debouncedSearch || undefined,
        role: roleFilter,
        status: statusFilter,
      });

      const userList = res.data || [];
      setUsers(userList);
      if (res.meta) {
        setMeta(res.meta);
      }

      // If drawer is currently inspecting a user, keep the live data updated
      if (inspectUser) {
        const refreshed = userList.find((u) => u.id === inspectUser.id);
        if (refreshed) setInspectUser(refreshed);
      }
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [cursor, debouncedSearch, roleFilter, statusFilter]);

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  // Action Handlers
  const handleConfirmRoleChange = async (userId: string, newRole: AdminUser["role"]) => {
    try {
      await adminService.updateUserRole(userId, { role: newRole });
      notify.success(t("customers.toast.roleUpdated"));
      await fetchUsers();
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("customers.toast.roleFailed")));
      throw err;
    }
  };

  const handleConfirmStatusChange = async (userId: string, status: AdminUser["status"], suspendedUntil: string | null) => {
    try {
      await adminService.updateUserStatus(userId, { status, suspendedUntil });
      notify.success(status === "active" ? t("customers.toast.reactivated") : t("customers.toast.suspended"));
      await fetchUsers();
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("customers.toast.statusFailed")));
      throw err;
    }
  };

  // Status column control: suspending needs a duration (modal), reactivating is immediate.
  const handleStatusSelect = (user: AdminUser, next: AdminUser["status"]) => {
    if (next === user.status) return;
    if (next === "suspended") {
      setSuspendModalUser(user);
    } else {
      handleConfirmStatusChange(user.id, "active", null).catch(() => {});
    }
  };

  const handleConfirmRevokeSessions = async (userId: string) => {
    try {
      await adminService.revokeUserSessions(userId);
      notify.success(t("customers.toast.sessionsRevoked"));
      await fetchUsers();
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("customers.toast.sessionsFailed")));
      throw err;
    }
  };

  // Role label
  const renderRoleBadge = (role: AdminUser["role"]) => (
    <span className="text-sm text-foreground">{t(`roles.${role}`)}</span>
  );

  if (loadError) {
    return <ErrorState onRetry={fetchUsers} />;
  }

  return (
    <div className="space-y-5">
      <PageHeader
        title={t("customers.title")}
        actions={
          <button
            onClick={() => fetchUsers()}
            disabled={loading}
            className="rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground transition-colors hover:bg-secondary disabled:opacity-50"
          >
            {t("customers.refresh")}
          </button>
        }
      />

      <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
        <div className="relative w-full sm:max-w-xs">
          <Search className="absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            placeholder={t("customers.searchPlaceholder")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full rounded-md border border-border bg-background py-2 pe-3 ps-9 text-sm text-foreground placeholder:text-muted-foreground focus:outline-hidden focus:ring-2 focus:ring-ring/30"
          />
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <CustomSelect
	value={roleFilter}
	onChange={(value) => {
		setRoleFilter(value);
		reset();
	}}
	options={["all", "customer", "seller", "admin", "superadmin"].map((r) => ({
		value: r,
		label: t(`customers.roles.${r}`),
	}))}
	ariaLabel={t("customers.colRole")}
/>

          <CustomSelect
	value={statusFilter}
	onChange={(value) => {
		setStatusFilter(value);
		reset();
	}}
	options={[
		{ value: "all", label: t("customers.allStatuses") },
		{ value: "active", label: t("customers.activeOnly") },
		{ value: "suspended", label: t("customers.suspendedOnly") },
	]}
	ariaLabel={t("customers.colStatus")}
/>
        </div>
      </div>

      <div className="overflow-hidden rounded-lg border border-border bg-background">
        <div className="overflow-x-auto">
          <ResponsiveTable className="w-full text-start text-sm md:min-w-max">
            <thead className="text-xs text-muted-foreground">
              <tr className="border-b border-border">
                <th className="py-2.5 ps-4 text-start font-medium">{t("customers.colIdentity")}</th>
                <th className="py-2.5 text-start font-medium">{t("customers.colRole")}</th>
                <th className="py-2.5 text-start font-medium">{t("customers.colStatus")}</th>
                <th className="py-2.5 text-start font-medium">{t("customers.colVerification")}</th>
                <th className="py-2.5 text-start font-medium">{t("customers.colJoined")}</th>
                <th className="py-2.5 pe-4 text-end font-medium">{t("customers.colActions")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {loading && users.length === 0 ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <tr key={i} className="animate-pulse">
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-3">
                        <div className="size-8 rounded-full bg-muted/60" />
                        <div className="space-y-1.5">
                          <div className="h-3 w-32 rounded-sm bg-muted/60" />
                          <div className="h-2.5 w-44 rounded-sm bg-muted/40" />
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3"><div className="h-4 w-16 rounded-full bg-muted/60" /></td>
                    <td className="px-4 py-3"><div className="h-4 w-14 rounded-full bg-muted/60" /></td>
                    <td className="px-4 py-3"><div className="h-4 w-16 rounded-sm bg-muted/40" /></td>
                    <td className="px-4 py-3"><div className="h-3 w-20 rounded-sm bg-muted/40" /></td>
                    <td className="px-4 py-3 text-end"><div className="ms-auto h-6 w-16 rounded-md bg-muted/40" /></td>
                  </tr>
                ))
              ) : users.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-12 text-center">
                    <p className="text-sm font-medium text-foreground">{t("customers.emptyTitle")}</p>
                    <p className="mt-1 text-sm text-muted-foreground">{t("customers.emptySub")}</p>
                  </td>
                </tr>
              ) : (
                users.map((u) => {
                  const isSelf = currentUser?.id && u.id && String(currentUser.id) === String(u.id);

                  const callerRank = ROLE_RANKS[currentUser?.role?.toLowerCase() || "customer"] || 1;
                  const targetRank = ROLE_RANKS[u.role?.toLowerCase() || "customer"] || 1;
                  // Hierarchy rule: Caller can only manage users strictly lower in rank
                  const canManageUser = !isSelf && callerRank > targetRank;

                  return (
                    <tr
                      key={u.id}
                      onClick={() => setInspectUser(u)}
                      className="cursor-pointer transition-colors hover:bg-secondary/30"
                    >
                      {/* Identity */}
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-3">
                          <div className="flex size-8 shrink-0 items-center justify-center rounded-full border border-border bg-secondary text-xs font-medium text-foreground">
                            {u.fullName ? u.fullName.slice(0, 2).toUpperCase() : u.email.slice(0, 2).toUpperCase()}
                          </div>
                          <div className="flex min-w-0 flex-col">
                            <div className="flex items-center gap-1.5">
                              <span className="truncate font-medium text-foreground">
                                {u.fullName || t("customers.unnamed")}
                              </span>
                              {isSelf ? <StatusPill>{t("customers.you")}</StatusPill> : null}
                            </div>
                            <span className="truncate font-mono text-xs text-muted-foreground">
                              {u.email}
                            </span>
                          </div>
                        </div>
                      </td>

                      {/* Role Badge */}
                      <td className="whitespace-nowrap px-4 py-3">
                        {renderRoleBadge(u.role)}
                      </td>

                      {/* Status */}
                      <td className="whitespace-nowrap px-4 py-3" onClick={(e) => e.stopPropagation()}>
                        {canManageUser ? (
                          <CustomSelect
	value={u.status}
	onChange={(value) => handleStatusSelect(u, value as AdminUser["status"])}
	options={[
		{ value: "active", label: t("status.active") },
		{ value: "suspended", label: t("status.suspended") },
	]}
	size="sm"
/>
                        ) : (
                          <StatusPill tone={statusTone(u.status)}>
                            {t(`status.${String(u.status).toLowerCase()}`) || u.status}
                          </StatusPill>
                        )}
                      </td>

                      {/* Email Verified */}
                      <td className="whitespace-nowrap px-4 py-3 text-xs text-muted-foreground">
                        {u.isEmailVerified ? t("status.verified") : t("status.unverified")}
                      </td>

                      {/* Joined Date */}
                      <td className="whitespace-nowrap px-4 py-3 text-xs text-muted-foreground">
                        {new Date(u.createdAt).toLocaleDateString(locale, {
                          year: "numeric",
                          month: "short",
                          day: "numeric",
                        })}
                      </td>

                      {/* Row Action Buttons */}
                      <td className="whitespace-nowrap px-4 py-3 text-end" onClick={(e) => e.stopPropagation()}>
                        <div className="inline-flex items-center gap-1">
                          <button
                            onClick={() => setInspectUser(u)}
                            className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                          >
                            {t("customers.inspect")}
                          </button>

                          {canManageUser ? (
                            <>
                              {currentUser?.role === "superadmin" ? (
                                <button
                                  onClick={() => setRoleModalUser(u)}
                                  className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                                >
                                  {t("customers.changeRole")}
                                </button>
                              ) : null}

                              <button
                                onClick={() => setRevokeModalUser(u)}
                                className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                              >
                                {t("customers.revokeSessions")}
                              </button>

                              {currentUser?.role === "superadmin" ? (
                                <button
                                  onClick={() => {
                                    if (confirm(t("customers.confirmDelete", { email: u.email }))) {
                                      adminService.deleteUser(u.id).then(() => {
                                        notify.success(t("customers.toast.deleted"));
                                        fetchUsers();
                                      });
                                    }
                                  }}
                                  className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                                >
                                  {t("customers.deleteUser")}
                                </button>
                              ) : null}
                            </>
                          ) : null}
                        </div>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </ResponsiveTable>
        </div>

        <CursorPager
          hasPrev={hasPrev}
          hasMore={meta.hasMore}
          page={page}
          onPrev={prev}
          onNext={() => next(meta.nextCursor)}
          loading={loading}
          shown={users.length}
        />
      </div>

      {/* Slide-over Inspector Drawer */}
      <UserSlideOverDrawer
        user={inspectUser}
        currentUserId={currentUser?.id}
        currentUserRole={currentUser?.role}
        isOpen={Boolean(inspectUser)}
        onClose={() => setInspectUser(null)}
        onChangeRole={(u) => setRoleModalUser(u)}
        onSuspend={(u) => setSuspendModalUser(u)}
        onRevokeSessions={(u) => setRevokeModalUser(u)}
      />

      {/* Change Role Modal */}
      {roleModalUser && (
        <ChangeRoleModal
          user={roleModalUser}
          callerRole={currentUser?.role}
          isOpen={Boolean(roleModalUser)}
          onClose={() => setRoleModalUser(null)}
          onConfirm={handleConfirmRoleChange}
        />
      )}

      {/* Suspend User Modal */}
      {suspendModalUser && (
        <SuspendUserModal
          user={suspendModalUser}
          isOpen={Boolean(suspendModalUser)}
          onClose={() => setSuspendModalUser(null)}
          onConfirm={handleConfirmStatusChange}
        />
      )}

      {/* Revoke Sessions Modal */}
      {revokeModalUser && (
        <RevokeSessionsModal
          user={revokeModalUser}
          isOpen={Boolean(revokeModalUser)}
          onClose={() => setRevokeModalUser(null)}
          onConfirm={handleConfirmRevokeSessions}
        />
      )}
    </div>
  );
}
