"use client";

import React, { useState, useEffect, useCallback } from "react";
import { 
  Search, 
  RefreshCw, 
  Shield, 
  MoreVertical, 
  UserX, 
  LogOut, 
  Users, 
  Filter,
  Eye
} from "lucide-react";
import { adminService, AdminUser } from "@/services/admin";
import { CursorPager, useCursorPagination } from "@/components/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/lib/pagination";
import { UserSlideOverDrawer } from "../components/UserSlideOverDrawer";
import { ChangeRoleModal } from "../components/ChangeRoleModal";
import { SuspendUserModal } from "../components/SuspendUserModal";
import { RevokeSessionsModal } from "../components/RevokeSessionsModal";
import { useAuthContext } from "@/context/AuthContext";
import { useTranslations } from "next-intl";
import { notify, apiErrorMessage } from "@/lib/notify";
import { ErrorState } from "@/components/ui/errorState";
import { ROLE_RANKS } from "@/lib/roles";

export function AdminCustomers() {
  const { user: currentUser } = useAuthContext();
  const t = useTranslations("admin");
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
    <span className="text-foreground">{t(`roles.${role}`)}</span>
  );

  if (loadError) {
    return <ErrorState onRetry={fetchUsers} />;
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-3xl border border-border/80 shadow-xs">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="text-2xl font-bold tracking-tight text-foreground">{t("customers.title")}</h1>
          </div>
          <p className="text-xs text-muted-foreground mt-1">
            {t("customers.subtitle")}
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => fetchUsers()}
            disabled={loading}
            className="p-2.5 rounded-xl border border-border bg-card hover:bg-secondary text-foreground text-xs font-semibold transition-colors flex items-center gap-2 disabled:opacity-50"
            title={t("customers.refreshTitle")}
          >
            <RefreshCw className={`size-3.5 ${loading ? "animate-spin" : ""}`} />
            <span className="hidden sm:inline">{t("customers.refresh")}</span>
          </button>
        </div>
      </div>

      {/* Control & Filter Bar */}
      <div className="bg-background border border-border/80 rounded-2xl p-4 shadow-xs flex flex-col md:flex-row gap-3 items-stretch md:items-center justify-between">
        {/* Search */}
        <div className="relative flex-1 max-w-md">
          <Search className="absolute start-3.5 top-2.5 size-4 text-muted-foreground" />
          <input
            type="text"
            placeholder={t("customers.searchPlaceholder")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full ps-10 pe-4 py-2 bg-secondary/20 border border-border rounded-xl text-xs text-foreground placeholder:text-muted-foreground focus:ring-2 focus:ring-primary focus:outline-hidden"
          />
        </div>

        {/* Filters */}
        <div className="flex flex-wrap items-center gap-2">
          {/* Role Filter */}
          <div className="flex items-center gap-1.5 bg-secondary/30 border border-border rounded-xl p-1 text-xs">
            {["all", "customer", "vendor", "admin", "superadmin"].map((r) => (
              <button
                key={r}
                onClick={() => {
                  setRoleFilter(r);
                  reset();
                }}
                className={`px-2.5 py-1 rounded-lg font-medium text-[11px] capitalize transition-all ${
                  roleFilter === r
                    ? "bg-card text-foreground font-bold shadow-xs border border-border"
                    : "text-muted-foreground hover:text-foreground"
                }`}
              >
                {t(`customers.roles.${r}`)}
              </button>
            ))}
          </div>

          {/* Status Filter */}
          <select
            value={statusFilter}
            onChange={(e) => {
              setStatusFilter(e.target.value);
              reset();
            }}
            className="py-1.5 px-3 bg-secondary/30 border border-border rounded-xl text-xs text-foreground font-medium focus:outline-hidden"
          >
            <option value="all">{t("customers.allStatuses")}</option>
            <option value="active">{t("customers.activeOnly")}</option>
            <option value="suspended">{t("customers.suspendedOnly")}</option>
          </select>
        </div>
      </div>

      {/* Dense User Table */}
      <div className="bg-background border border-border/80 rounded-2xl overflow-hidden shadow-xs">
        <div className="overflow-x-auto">
          <table className="w-full text-start border-collapse text-xs">
            <thead>
              <tr className="border-b border-border/80 bg-muted/20 text-muted-foreground uppercase text-[10px] font-bold tracking-wider">
                <th className="py-3 px-4">{t("customers.colIdentity")}</th>
                <th className="py-3 px-4">{t("customers.colRole")}</th>
                <th className="py-3 px-4">{t("customers.colStatus")}</th>
                <th className="py-3 px-4">{t("customers.colVerification")}</th>
                <th className="py-3 px-4">{t("customers.colJoined")}</th>
                <th className="py-3 px-4 text-end">{t("customers.colActions")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {loading && users.length === 0 ? (
                // Skeletons
                Array.from({ length: 5 }).map((_, i) => (
                  <tr key={i} className="animate-pulse">
                    <td className="py-3.5 px-4">
                      <div className="flex items-center gap-3">
                        <div className="size-8 rounded-full bg-muted/60" />
                        <div className="space-y-1.5">
                          <div className="h-3 w-32 bg-muted/60 rounded-sm" />
                          <div className="h-2.5 w-44 bg-muted/40 rounded-sm" />
                        </div>
                      </div>
                    </td>
                    <td className="py-3.5 px-4"><div className="h-4 w-16 bg-muted/60 rounded-full" /></td>
                    <td className="py-3.5 px-4"><div className="h-4 w-14 bg-muted/60 rounded-full" /></td>
                    <td className="py-3.5 px-4"><div className="h-4 w-16 bg-muted/40 rounded-sm" /></td>
                    <td className="py-3.5 px-4"><div className="h-3 w-20 bg-muted/40 rounded-sm" /></td>
                    <td className="py-3.5 px-4 text-end"><div className="h-6 w-16 bg-muted/40 rounded-lg ms-auto" /></td>
                  </tr>
                ))
              ) : users.length === 0 ? (
                <tr>
                  <td colSpan={6} className="py-12 text-center text-muted-foreground">
                    <Users className="size-8 mx-auto mb-2 text-muted-foreground/50" />
                    <p className="font-semibold text-foreground text-sm">{t("customers.emptyTitle")}</p>
                    <p className="text-xs mt-0.5">{t("customers.emptySub")}</p>
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
                      className="hover:bg-muted/15 transition-colors cursor-pointer group"
                    >
                      {/* Identity */}
                      <td className="py-3 px-4">
                        <div className="flex items-center gap-3">
                          <div className="size-8 rounded-full bg-primary/10 border border-primary/20 text-primary flex items-center justify-center font-bold text-xs shrink-0">
                            {u.fullName ? u.fullName.slice(0, 2).toUpperCase() : u.email.slice(0, 2).toUpperCase()}
                          </div>
                          <div className="flex flex-col min-w-0">
                            <div className="flex items-center gap-1.5">
                              <span className="font-bold text-foreground truncate group-hover:text-primary transition-colors">
                                {u.fullName || t("customers.unnamed")}
                              </span>
                              {isSelf && (
                                <span className="px-1.5 py-0.5 rounded text-[9px] font-bold bg-accent-brand/10 text-accent-brand border border-accent-brand/20">
                                  {t("customers.you")}
                                </span>
                              )}
                            </div>
                            <span className="text-[11px] font-mono text-muted-foreground truncate">
                              {u.email}
                            </span>
                          </div>
                        </div>
                      </td>

                      {/* Role Badge */}
                      <td className="py-3 px-4 whitespace-nowrap">
                        {renderRoleBadge(u.role)}
                      </td>

                      {/* Status */}
                      <td className="py-3 px-4 whitespace-nowrap" onClick={(e) => e.stopPropagation()}>
                        {canManageUser ? (
                          <select
                            value={u.status}
                            onChange={(e) => handleStatusSelect(u, e.target.value as AdminUser["status"])}
                            className="px-2.5 py-1 bg-secondary/50 border border-border rounded-lg text-xs font-semibold text-foreground focus:outline-hidden cursor-pointer"
                          >
                            <option value="active">{t("status.active")}</option>
                            <option value="suspended">{t("status.suspended")}</option>
                          </select>
                        ) : (
                          <span className="text-foreground">
                            {t(`status.${String(u.status).toLowerCase()}`) || u.status}
                          </span>
                        )}
                      </td>

                      {/* Email Verified */}
                      <td className="py-3 px-4 whitespace-nowrap text-foreground">
                        {u.isEmailVerified ? t("status.verified") : t("status.unverified")}
                      </td>

                      {/* Joined Date */}
                      <td className="py-3 px-4 whitespace-nowrap text-muted-foreground font-mono text-[11px]">
                        {new Date(u.createdAt).toLocaleDateString(undefined, {
                          year: "numeric",
                          month: "short",
                          day: "numeric",
                        })}
                      </td>

                      {/* Row Action Buttons */}
                      <td className="py-3 px-4 text-end whitespace-nowrap" onClick={(e) => e.stopPropagation()}>
                        <div className="flex items-center justify-end gap-1.5">
                          <button
                            onClick={() => setInspectUser(u)}
                            className="p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors"
                            title={t("customers.inspect")}
                          >
                            <Eye className="size-4" />
                          </button>

                          {canManageUser && (
                            <>
                              <button
                                onClick={() => setRoleModalUser(u)}
                                className="p-1.5 rounded-lg text-purple-600 hover:text-purple-700 hover:bg-purple-500/10 transition-colors"
                                title={t("customers.changeRole")}
                              >
                                <Shield className="size-4" />
                              </button>

                              <button
                                onClick={() => setRevokeModalUser(u)}
                                className="p-1.5 rounded-lg text-amber-600 hover:bg-amber-500/10 transition-colors"
                                title={t("customers.revokeSessions")}
                              >
                                <LogOut className="size-4" />
                              </button>

                              {currentUser?.role === "superadmin" && (
                                <button
                                  onClick={() => {
                                    if (confirm(t("customers.confirmDelete", { email: u.email }))) {
                                      adminService.deleteUser(u.id).then(() => {
                                        notify.success(t("customers.toast.deleted"));
                                        fetchUsers();
                                      });
                                    }
                                  }}
                                  className="p-1.5 rounded-lg text-destructive hover:bg-destructive/10 transition-colors"
                                  title={t("customers.deleteUser")}
                                >
                                  <UserX className="size-4" />
                                </button>
                              )}
                            </>
                          )}
                        </div>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination Bar */}
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
