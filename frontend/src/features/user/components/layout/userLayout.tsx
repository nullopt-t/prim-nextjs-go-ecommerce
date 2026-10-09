"use client";

import { useEffect } from "react";
import { useAuthContext } from "@/context/AuthContext";
import { useRouter, usePathname } from "@/i18n/navigation";
import UserSideBar from "@/features/user/components/ui/sideBar";
import MobileNav from "@/features/user/components/ui/mobileNav";
import { Loader2 } from "lucide-react";

export default function UserLayout({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, loading } = useAuthContext();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (!loading && !isAuthenticated) {
      const returnUrl = encodeURIComponent(pathname);
      router.replace(`/auth?redirect=${returnUrl}`);
    }
  }, [loading, isAuthenticated, router, pathname]);

  if (loading) {
    return (
      <div className="min-h-[60vh] flex flex-col items-center justify-center gap-3">
        <Loader2 className="w-8 h-8 animate-spin text-accent-brand" />
        <p className="text-sm text-muted-foreground animate-pulse">Loading your account...</p>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return (
    <>
      <div className="w-full px-3 sm:px-6 lg:px-8 py-4 sm:py-6 pb-24 md:pb-8">
        {/* Desktop: sidebar + content */}
        <div className="flex flex-row gap-6">
          <UserSideBar />
          <main className="flex-1 min-w-0">{children}</main>
        </div>
      </div>
      {/* Fixed bottom nav — mobile only */}
      <MobileNav />
    </>
  );
}

