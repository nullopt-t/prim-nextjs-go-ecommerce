"use client";

import React, { useState, useEffect, useCallback } from "react";
import { 
  Search, 
  RefreshCw, 
  Shield, 
  CheckCircle2, 
  XCircle, 
  MoreVertical, 
  UserX, 
  UserCheck, 
  LogOut, 
  ShieldAlert, 
  Users, 
  ChevronLeft, 
  ChevronRight,
  Filter,
  Eye,
  Store,
  Headphones,
  ShieldCheck
} from "lucide-react";
import { adminService, AdminUser, PaginationMeta } from "@/services/admin";
import { UserSlideOverDrawer } from "../components/UserSlideOverDrawer";
import { ChangeRoleModal } from "../components/ChangeRoleModal";
import { SuspendUserModal } from "../components/SuspendUserModal";
import { RevokeSessionsModal } from "../components/RevokeSessionsModal";
import { useAuthContext } from "@/context/AuthContext";
import { toast } from "sonner";
import { ROLE_RANKS } from "@/lib/roles";

export function AdminCustomers() {
  const { user: currentUser } = useAuthContext();
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [meta, setMeta] = useState<PaginationMeta>({
    page: 1,
    pageSize: 10,
    totalItems: 0,
    totalPages: 1,
    hasPrevious: false,
    hasNext: false,
  });

  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [roleFilter, setRoleFilter] = useState("all");
  const [statusFilter, setStatusFilter] = useState("all");
  const [page, setPage] = useState(1);

  // Active User Drawer & Modals state
  const [inspectUser, setInspectUser] = useState<AdminUser | null>(null);
  const [roleModalUser, setRoleModalUser] = useState<AdminUser | null>(null);
  const [suspendModalUser, setSuspendModalUser] = useState<AdminUser | null>(null);
  const [revokeModalUser, setRevokeModalUser] = useState<AdminUser | null>(null);

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setPage(1); // reset to first page on search
    }, 350);
    return () => clearTimeout(timer);
  }, [search]);

  const fetchUsers = useCallback(async () => {
    setLoading(true);
    try {
      const res = await adminService.getUsers({
        page,
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
    } catch (err: any) {
      toast.error(err.message || "Failed to load user records");
    } finally {
      setLoading(false);
    }
  }, [page, debouncedSearch, roleFilter, statusFilter]);

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  // Action Handlers
  const handleConfirmRoleChange = async (userId: string, newRole: AdminUser["role"]) => {
    try {
      await adminService.updateUserRole(userId, { role: newRole });
      toast.success("User role updated successfully");
      await fetchUsers();
    } catch (err: any) {
      toast.error(err.message || "Failed to update role");
      throw err;
    }
  };

  const handleConfirmStatusChange = async (userId: string, status: AdminUser["status"], suspendedUntil: string | null) => {
    try {
      await adminService.updateUserStatus(userId, { status, suspendedUntil });
      toast.success(status === "active" ? "User account reactivated" : "User account suspended");
      await fetchUsers();
    } catch (err: any) {
      toast.error(err.message || "Failed to update account status");
      throw err;
    }
  };

  const handleConfirmRevokeSessions = async (userId: string) => {
    try {
      await adminService.revokeUserSessions(userId);
      toast.success("All active user sessions revoked successfully");
      await fetchUsers();
    } catch (err: any) {
      toast.error(err.message || "Failed to revoke sessions");
      throw err;
    }
  };

  // Helper badge stylers
  const renderRoleBadge = (role: AdminUser["role"]) => {
    switch (role) {
      case "superadmin":
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-bold bg-purple-500/15 text-purple-700 dark:text-purple-300 border border-purple-500/30">
            <ShieldAlert className="size-3 text-purple-600 dark:text-purple-400" />
            Superadmin
          </span>
        );
      case "admin":
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-bold bg-blue-500/15 text-blue-700 dark:text-blue-300 border border-blue-500/30">
            <ShieldCheck className="size-3 text-blue-600 dark:text-blue-400" />
            Admin
          </span>
        );
      case "vendor":
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-bold bg-amber-500/15 text-amber-700 dark:text-amber-300 border border-amber-500/30">
            <Store className="size-3 text-amber-600 dark:text-amber-400" />
            Vendor
          </span>
        );
      case "support":
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-bold bg-teal-500/15 text-teal-700 dark:text-teal-300 border border-teal-500/30">
            <Headphones className="size-3 text-teal-600 dark:text-teal-400" />
            Support
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-secondary text-secondary-foreground border border-border">
            Customer
          </span>
        );
    }
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-3xl border border-border/80 shadow-xs">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="text-2xl font-bold tracking-tight text-foreground">User & Staff Management</h1>
            <span className="px-2.5 py-0.5 rounded-full text-[11px] font-bold bg-primary/10 text-primary border border-primary/20">
              {meta.totalItems} Total Accounts
            </span>
          </div>
          <p className="text-xs text-muted-foreground mt-1">
            Superadmin directory: inspect accounts, assign administrative roles, manage suspensions, and invalidate credentials.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => fetchUsers()}
            disabled={loading}
            className="p-2.5 rounded-xl border border-border bg-card hover:bg-secondary text-foreground text-xs font-semibold transition-colors flex items-center gap-2 disabled:opacity-50"
            title="Refresh User Directory"
          >
            <RefreshCw className={`size-3.5 ${loading ? "animate-spin" : ""}`} />
            <span className="hidden sm:inline">Refresh</span>
          </button>
        </div>
      </div>

      {/* Control & Filter Bar */}
      <div className="bg-background border border-border/80 rounded-2xl p-4 shadow-xs flex flex-col md:flex-row gap-3 items-stretch md:items-center justify-between">
        {/* Search */}
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3.5 top-2.5 size-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search by email, name, or UUID..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-secondary/20 border border-border rounded-xl text-xs text-foreground placeholder:text-muted-foreground focus:ring-2 focus:ring-primary focus:outline-hidden"
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
                  setPage(1);
                }}
                className={`px-2.5 py-1 rounded-lg font-medium text-[11px] capitalize transition-all ${
                  roleFilter === r
                    ? "bg-card text-foreground font-bold shadow-xs border border-border"
                    : "text-muted-foreground hover:text-foreground"
                }`}
              >
                {r}
              </button>
            ))}
          </div>

          {/* Status Filter */}
          <select
            value={statusFilter}
            onChange={(e) => {
              setStatusFilter(e.target.value);
              setPage(1);
            }}
            className="py-1.5 px-3 bg-secondary/30 border border-border rounded-xl text-xs text-foreground font-medium focus:outline-hidden"
          >
            <option value="all">All Statuses</option>
            <option value="active">Active Only</option>
            <option value="suspended">Suspended Only</option>
          </select>
        </div>
      </div>

      {/* Dense User Table */}
      <div className="bg-background border border-border/80 rounded-2xl overflow-hidden shadow-xs">
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse text-xs">
            <thead>
              <tr className="border-b border-border/80 bg-muted/20 text-muted-foreground uppercase text-[10px] font-bold tracking-wider">
                <th className="py-3 px-4">User Identity</th>
                <th className="py-3 px-4">Permission Role</th>
                <th className="py-3 px-4">Status</th>
                <th className="py-3 px-4">Email Verification</th>
                <th className="py-3 px-4">Joined Date</th>
                <th className="py-3 px-4 text-right">Actions</th>
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
                    <td className="py-3.5 px-4 text-right"><div className="h-6 w-16 bg-muted/40 rounded-lg ml-auto" /></td>
                  </tr>
                ))
              ) : users.length === 0 ? (
                <tr>
                  <td colSpan={6} className="py-12 text-center text-muted-foreground">
                    <Users className="size-8 mx-auto mb-2 text-muted-foreground/50" />
                    <p className="font-semibold text-foreground text-sm">No accounts found</p>
                    <p className="text-xs mt-0.5">Try adjusting your filters or search terms.</p>
                  </td>
                </tr>
              ) : (
                users.map((u) => {
                  const isSuspended = u.status === "suspended";
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
                                {u.fullName || "Unnamed User"}
                              </span>
                              {isSelf && (
                                <span className="px-1.5 py-0.5 rounded text-[9px] font-bold bg-accent-brand/10 text-accent-brand border border-accent-brand/20">
                                  You
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
                      <td className="py-3 px-4 whitespace-nowrap">
                        <span className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-bold border ${
                          u.status === "active"
                            ? "bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border-emerald-500/30"
                            : "bg-rose-500/10 text-rose-700 dark:text-rose-300 border-rose-500/30"
                        }`}>
                          <span className={`size-1.5 rounded-full ${u.status === "active" ? "bg-emerald-500" : "bg-rose-500"}`} />
                          {u.status}
                        </span>
                      </td>

                      {/* Email Verified */}
                      <td className="py-3 px-4 whitespace-nowrap">
                        {u.isEmailVerified ? (
                          <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-emerald-600 dark:text-emerald-400">
                            <CheckCircle2 className="size-3.5" /> Verified
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-muted-foreground">
                            <XCircle className="size-3.5 text-muted-foreground/60" /> Unverified
                          </span>
                        )}
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
                      <td className="py-3 px-4 text-right whitespace-nowrap" onClick={(e) => e.stopPropagation()}>
                        <div className="flex items-center justify-end gap-1.5">
                          <button
                            onClick={() => setInspectUser(u)}
                            className="p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors"
                            title="Inspect User Details"
                          >
                            <Eye className="size-4" />
                          </button>

                          {canManageUser && (
                            <>
                              <button
                                onClick={() => setRoleModalUser(u)}
                                className="p-1.5 rounded-lg text-purple-600 hover:text-purple-700 hover:bg-purple-500/10 transition-colors"
                                title="Change Role"
                              >
                                <Shield className="size-4" />
                              </button>

                              <button
                                onClick={() => setSuspendModalUser(u)}
                                className={`p-1.5 rounded-lg transition-colors ${
                                  isSuspended
                                    ? "text-emerald-600 hover:bg-emerald-500/10"
                                    : "text-rose-600 hover:bg-rose-500/10"
                                }`}
                                title={isSuspended ? "Reactivate Account" : "Suspend Account"}
                              >
                                {isSuspended ? <UserCheck className="size-4" /> : <UserX className="size-4" />}
                              </button>

                              <button
                                onClick={() => setRevokeModalUser(u)}
                                className="p-1.5 rounded-lg text-amber-600 hover:bg-amber-500/10 transition-colors"
                                title="Revoke Sessions"
                              >
                                <LogOut className="size-4" />
                              </button>

                              {currentUser?.role === "superadmin" && (
                                <button
                                  onClick={() => {
                                    if (confirm(`Delete user ${u.email}? This cannot be undone.`)) {
                                      adminService.deleteUser(u.id).then(() => {
                                        toast.success("User deleted");
                                        fetchUsers();
                                      });
                                    }
                                  }}
                                  className="p-1.5 rounded-lg text-destructive hover:bg-destructive/10 transition-colors"
                                  title="Delete User"
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
        <div className="p-4 border-t border-border/80 bg-muted/10 flex items-center justify-between text-xs text-muted-foreground">
          <span>
            Showing <strong className="text-foreground">{users.length}</strong> of{" "}
            <strong className="text-foreground">{meta.totalItems}</strong> accounts
          </span>

          <div className="flex items-center gap-2">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={!meta.hasPrevious || loading}
              className="p-1.5 rounded-lg border border-border bg-card hover:bg-secondary text-foreground disabled:opacity-40 transition-colors"
              title="Previous Page"
            >
              <ChevronLeft className="size-4" />
            </button>
            <span className="font-mono text-xs px-2 font-bold text-foreground">
              Page {meta.page} of {Math.max(1, meta.totalPages)}
            </span>
            <button
              onClick={() => setPage((p) => p + 1)}
              disabled={!meta.hasNext || loading}
              className="p-1.5 rounded-lg border border-border bg-card hover:bg-secondary text-foreground disabled:opacity-40 transition-colors"
              title="Next Page"
            >
              <ChevronRight className="size-4" />
            </button>
          </div>
        </div>
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
