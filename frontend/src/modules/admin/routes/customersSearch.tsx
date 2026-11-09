"use client";

import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { useState, useEffect, useCallback, useRef } from "react";
import { useSearchParams } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { adminService, type AdminUser } from "@/modules/admin/api/adminService";
import { useAuthContext } from "@/modules/auth/model/authContext";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill } from "@/shared/ui/statusPill";
import { CursorPager, useCursorPagination } from "@/shared/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/shared/lib/pagination";
import { UserSlideOverDrawer } from "../components/UserSlideOverDrawer";
import { ChangeRoleModal } from "../components/ChangeRoleModal";
import { SuspendUserModal } from "../components/SuspendUserModal";
import { RevokeSessionsModal } from "../components/RevokeSessionsModal";
import { Link } from "@/shared/i18n/navigation";

const KNOWN_ROLES = ["all", "customer", "seller", "admin", "superadmin"];
const KNOWN_STATUSES = ["all", "active", "suspended"];

function initials(u: AdminUser): string {
	return u.fullName ? u.fullName.slice(0, 2).toUpperCase() : u.email.slice(0, 2).toUpperCase();
}

export function AdminCustomersSearch() {
	const searchParams = useSearchParams();
	const initialRole = (() => {
		const r = (searchParams.get("role") || "all").toLowerCase();
		return KNOWN_ROLES.includes(r) ? r : "all";
	})();
	const initialStatus = (() => {
		const s = (searchParams.get("status") || "all").toLowerCase();
		return KNOWN_STATUSES.includes(s) ? s : "all";
	})();

	const t = useTranslations("admin");
	const locale = useLocale();
	const { user: currentUser } = useAuthContext();

	const [users, setUsers] = useState<AdminUser[]>([]);
	const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState(false);

	const [searchQuery, setSearchQuery] = useState("");
	const [debouncedSearch, setDebouncedSearch] = useState("");
	const [roleFilter, setRoleFilter] = useState(initialRole);
	const [statusFilter, setStatusFilter] = useState(initialStatus);

	const [roleOpen, setRoleOpen] = useState(false);
	const [statusOpen, setStatusOpen] = useState(false);
	const roleRef = useRef<HTMLDivElement>(null);
	const statusRef = useRef<HTMLDivElement>(null);

	const [inspectUser, setInspectUser] = useState<AdminUser | null>(null);
	const [roleModalUser, setRoleModalUser] = useState<AdminUser | null>(null);
	const [suspendModalUser, setSuspendModalUser] = useState<AdminUser | null>(null);
	const [revokeModalUser, setRevokeModalUser] = useState<AdminUser | null>(null);

	useEffect(() => {
		const id = setTimeout(() => setDebouncedSearch(searchQuery.trim()), 300);
		return () => clearTimeout(id);
	}, [searchQuery]);

	const handleSearchKey = (e: React.KeyboardEvent) => {
		if (e.key === "Enter") setDebouncedSearch(searchQuery.trim());
	};

	useEffect(() => {
		const onDown = (e: MouseEvent) => {
			if (!roleRef.current?.contains(e.target as Node)) setRoleOpen(false);
			if (!statusRef.current?.contains(e.target as Node)) setStatusOpen(false);
		};
		const onKey = (e: KeyboardEvent) => {
			if (e.key === "Escape") {
				setRoleOpen(false);
				setStatusOpen(false);
			}
		};
		document.addEventListener("mousedown", onDown);
		document.addEventListener("keydown", onKey);
		return () => {
			document.removeEventListener("mousedown", onDown);
			document.removeEventListener("keydown", onKey);
		};
	}, []);

	const { cursor, hasPrev, next, prev, reset, page } = useCursorPagination();

	const fetchUsers = useCallback(async () => {
		try {
			setLoading(true);
			setLoadError(false);
			const res = await adminService.getUsers({
				cursor: cursor || undefined,
				pageSize: 50,
				search: debouncedSearch || undefined,
				role: roleFilter !== "all" ? roleFilter : undefined,
				status: statusFilter !== "all" ? statusFilter : undefined,
			});
			setUsers(res.data ?? []);
			if (res.meta) setMeta(res.meta);
		} catch {
			setLoadError(true);
		} finally {
			setLoading(false);
		}
	}, [cursor, debouncedSearch, roleFilter, statusFilter]);

	useEffect(() => {
		setRoleFilter(initialRole);
		setStatusFilter(initialStatus);
		reset();
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [initialRole, initialStatus]);

	useEffect(() => {
		reset();
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [debouncedSearch, roleFilter, statusFilter]);

	useEffect(() => {
		fetchUsers();
	}, [fetchUsers]);

	const handleConfirmRoleChange = async (userId: string, newRole: AdminUser["role"]) => {
		try {
			await adminService.updateUserRole(userId, { role: newRole });
			notify.success(t("customers.toast.roleUpdated"));
			await fetchUsers();
		} catch (err: unknown) {
			notify.error(apiErrorMessage(err, t("customers.toast.roleFailed")));
			throw err;
		}
	};

	const handleConfirmStatusChange = async (userId: string, status: AdminUser["status"], suspendedUntil: string | null) => {
		try {
			await adminService.updateUserStatus(userId, { status, suspendedUntil });
			notify.success(status === "active" ? t("customers.toast.reactivated") : t("customers.toast.suspended"));
			await fetchUsers();
		} catch (err: unknown) {
			notify.error(apiErrorMessage(err, t("customers.toast.statusFailed")));
			throw err;
		}
	};

	const handleConfirmRevokeSessions = async (userId: string) => {
		try {
			await adminService.revokeUserSessions(userId);
			notify.success(t("customers.toast.sessionsRevoked"));
			await fetchUsers();
		} catch (err: unknown) {
			notify.error(apiErrorMessage(err, t("customers.toast.sessionsFailed")));
			throw err;
		}
	};

	const roleOptions = ["all", "customer", "seller", "admin", "superadmin"].map((r) => ({
		value: r,
		label: t(`customers.roles.${r}`),
	}));
	const statusOptions = [
		{ value: "all", label: t("customers.allStatuses") },
		{ value: "active", label: t("customers.activeOnly") },
		{ value: "suspended", label: t("customers.suspendedOnly") },
	];
	const roleLabel = roleOptions.find((o) => o.value === roleFilter)?.label ?? roleFilter;
	const statusLabel = statusOptions.find((o) => o.value === statusFilter)?.label ?? statusFilter;

	return (
		<div className="space-y-6">
			<div className="flex flex-col gap-3">
				<Link href="/admin/customers" className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors">
					{t("customers.backToCustomersOverview")}
				</Link>
				<PageHeader title={t("customers.searchTitle")} description={t("customers.searchSubtitle")} />
			</div>

			<div className="rounded-xl border border-border/70 bg-card p-4 sm:p-6 space-y-4">
				<div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
					<div className="space-y-1.5">
						<label htmlFor="customersSearchQuery" className="text-xs font-medium text-muted-foreground">
							{t("customers.searchLabel")}
						</label>
						<input
							id="customersSearchQuery"
							type="text"
							value={searchQuery}
							onChange={(e) => setSearchQuery(e.target.value)}
							onKeyDown={handleSearchKey}
							placeholder={t("customers.searchPlaceholderCustomers")}
							className="w-full h-11 px-3 rounded-xl border border-border bg-background text-sm text-foreground focus:outline-hidden focus:border-border hover:bg-muted/30 transition-colors"
						/>
					</div>

					<div className="space-y-1.5 relative" ref={roleRef}>
						<span id="customersSearchRoleLabel" className="text-xs font-medium text-muted-foreground">{t("customers.colRole")}</span>
						<button
							type="button"
							onClick={() => setRoleOpen((v) => !v)}
							aria-expanded={roleOpen}
							aria-haspopup="listbox"
							aria-labelledby="customersSearchRoleLabel"
							className="w-full h-11 px-3 rounded-xl border border-border bg-background text-sm text-foreground text-start hover:bg-muted/30 transition-colors cursor-pointer"
						>
							{roleLabel}
						</button>
						{roleOpen && (
							<div role="listbox" aria-label={t("customers.colRole")} className="absolute z-50 mt-2 w-full rounded-xl border border-border/70 bg-card py-1.5 shadow-lg overflow-hidden">
								{roleOptions.map((o) => (
									<button
										key={o.value}
										type="button"
										role="option"
										aria-selected={o.value === roleFilter}
										onClick={() => {
											setRoleFilter(o.value);
											setRoleOpen(false);
										}}
										className={`w-full px-3 py-2 text-start text-sm transition-colors cursor-pointer ${o.value === roleFilter ? "font-semibold text-foreground bg-muted/40" : "text-muted-foreground hover:text-foreground hover:bg-muted/30"}`}
									>
										{o.label}
									</button>
								))}
							</div>
						)}
					</div>

					<div className="space-y-1.5 relative" ref={statusRef}>
						<span id="customersSearchStatusLabel" className="text-xs font-medium text-muted-foreground">{t("customers.colStatus")}</span>
						<button
							type="button"
							onClick={() => setStatusOpen((v) => !v)}
							aria-expanded={statusOpen}
							aria-haspopup="listbox"
							aria-labelledby="customersSearchStatusLabel"
							className="w-full h-11 px-3 rounded-xl border border-border bg-background text-sm text-foreground text-start hover:bg-muted/30 transition-colors cursor-pointer"
						>
							{statusLabel}
						</button>
						{statusOpen && (
							<div role="listbox" aria-label={t("customers.colStatus")} className="absolute z-50 mt-2 w-full rounded-xl border border-border/70 bg-card py-1.5 shadow-lg overflow-hidden">
								{statusOptions.map((o) => (
									<button
										key={o.value}
										type="button"
										role="option"
										aria-selected={o.value === statusFilter}
										onClick={() => {
											setStatusFilter(o.value);
											setStatusOpen(false);
										}}
										className={`w-full px-3 py-2 text-start text-sm transition-colors cursor-pointer ${o.value === statusFilter ? "font-semibold text-foreground bg-muted/40" : "text-muted-foreground hover:text-foreground hover:bg-muted/30"}`}
									>
										{o.label}
									</button>
								))}
							</div>
						)}
					</div>
				</div>
			</div>

			{loadError ? (
				<ErrorState onRetry={fetchUsers} />
			) : (
				<div className="overflow-hidden rounded-xl border border-border/70 bg-card">
					<div className="overflow-x-auto">
						<ResponsiveTable className="w-full text-start text-sm">
							<thead className="text-xs text-muted-foreground bg-muted/30">
								<tr className="border-b border-border/70">
									<th className="py-3 ps-4 text-start font-medium">{t("customers.colIdentity")}</th>
									<th className="py-3 text-start font-medium">{t("customers.colRole")}</th>
									<th className="py-3 text-start font-medium">{t("customers.colStatus")}</th>
									<th className="py-3 text-start font-medium">{t("customers.colJoined")}</th>
									<th className="py-3 pe-4 text-end font-medium">{t("customers.colActions")}</th>
								</tr>
							</thead>
							<tbody className="divide-y divide-border/50">
								{loading ? (
									<tr>
										<td colSpan={5} className="py-12 text-center text-sm text-muted-foreground">{t("common.loading")}</td>
									</tr>
								) : users.length === 0 ? (
									<tr>
										<td colSpan={5} className="py-12 text-center">
											<p className="text-sm font-medium text-foreground">{t("customers.emptyTitle")}</p>
											<p className="mt-1 text-sm text-muted-foreground">{t("customers.emptySub")}</p>
										</td>
									</tr>
								) : (
									users.map((u) => (
										<tr key={u.id} className="hover:bg-muted/40 transition-colors">
											<td className="py-3.5 ps-4">
												<div className="flex items-center gap-3">
													<div className="flex size-8 shrink-0 items-center justify-center rounded-full border border-border bg-secondary text-xs font-medium text-foreground">
														{initials(u)}
													</div>
													<div className="min-w-0">
														<div className="truncate font-medium text-foreground">{u.fullName || t("customers.unnamed")}</div>
														<div className="truncate font-mono text-xs text-muted-foreground">{u.email}</div>
													</div>
												</div>
											</td>
											<td className="py-3.5">{t(`roles.${u.role}`) || t(`customers.roles.${u.role}`)}</td>
											<td className="py-3.5">
												<StatusPill>{t(`status.${String(u.status).toLowerCase()}`) || u.status}</StatusPill>
											</td>
											<td className="py-3.5 text-xs text-muted-foreground">
												{new Date(u.createdAt).toLocaleDateString(locale, { year: "numeric", month: "short", day: "numeric" })}
											</td>
											<td className="py-3.5 pe-4 text-end">
												<button
													type="button"
													onClick={() => setInspectUser(u)}
													className="text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
												>
													{t("customers.manageAction")}
												</button>
											</td>
										</tr>
									))
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
			)}

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

			{roleModalUser && (
				<ChangeRoleModal
					user={roleModalUser}
					callerRole={currentUser?.role}
					isOpen={Boolean(roleModalUser)}
					onClose={() => setRoleModalUser(null)}
					onConfirm={handleConfirmRoleChange}
				/>
			)}

			{suspendModalUser && (
				<SuspendUserModal
					user={suspendModalUser}
					isOpen={Boolean(suspendModalUser)}
					onClose={() => setSuspendModalUser(null)}
					onConfirm={handleConfirmStatusChange}
				/>
			)}

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
