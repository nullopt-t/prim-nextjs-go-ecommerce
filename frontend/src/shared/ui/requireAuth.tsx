"use client";

import { useEffect } from "react";
import { usePathname, useRouter } from "@/shared/i18n/navigation";
import { useAuthContext } from "@/modules/auth/model/authContext";

/**
 * Authentication-only UX guard. Merchant authorization remains a backend
 * resource check; this component only ensures a session exists before the
 * seller provider starts calling the profile endpoint.
 */
export function RequireAuth({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuthContext();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (loading || user) return;
    const destination = pathname || "/seller";
    router.replace(`/auth?redirect=${encodeURIComponent(destination)}`);
  }, [loading, pathname, router, user]);

  if (loading || !user) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-accent-brand border-t-transparent" />
      </div>
    );
  }

  return <>{children}</>;
}
