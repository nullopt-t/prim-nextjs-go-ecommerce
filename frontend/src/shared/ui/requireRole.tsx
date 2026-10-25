"use client";

import { useEffect } from "react";
import { useRouter } from "@/shared/i18n/navigation";
import { useAuthContext } from "@/modules/auth/model/authContext";
import { ROLE_RANKS } from "@/shared/lib/roles";

interface RequireRoleProps {
	minimumRole: keyof typeof ROLE_RANKS;
	children: React.ReactNode;
}

/**
 * Client-side UX guard for the admin and vendor consoles. The backend remains
 * the authority for authorization; this only prevents rendering a console the
 * current session cannot use.
 */
export function RequireRole({ minimumRole, children }: RequireRoleProps) {
	const { user, loading } = useAuthContext();
	const router = useRouter();

	const userRank = ROLE_RANKS[user?.role ?? "customer"] ?? 0;
	const requiredRank = ROLE_RANKS[minimumRole] ?? Number.MAX_SAFE_INTEGER;
	const allowed = Boolean(user) && userRank >= requiredRank;

	useEffect(() => {
		if (!loading && !allowed) {
			router.replace("/auth");
		}
	}, [allowed, loading, router]);

	if (loading || !allowed) {
		return (
			<div className="flex min-h-[60vh] items-center justify-center">
				<div className="size-8 animate-spin rounded-full border-2 border-accent-brand border-t-transparent" />
			</div>
		);
	}

	return <>{children}</>;
}
