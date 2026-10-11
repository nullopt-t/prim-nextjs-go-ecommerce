"use client";

import { useEffect } from "react";
import { useAuthContext } from "@/context/AuthContext";
import { useRouter, usePathname } from "@/i18n/navigation";
import UserTopTabs from "@/features/user/components/ui/userTopTabs";
import MobileNav from "@/features/user/components/ui/mobileNav";
import { Header, Footer } from "@/components/ui";
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
    <div className="flex flex-col min-h-screen bg-background text-foreground">
      <Header />
      <div className="w-full px-4 sm:px-6 lg:px-10 py-5 sm:py-8 pb-24 md:pb-12 flex-1">
        <UserTopTabs />
        <main className="w-full">{children}</main>
      </div>
      {/* Fixed bottom nav — mobile only */}
      <MobileNav />
      <Footer />
    </div>
  );
}

