"use client";

import { useEffect } from "react";
import { useAuthContext } from "@/context/AuthContext";
import { useRouter, usePathname } from "@/i18n/navigation";
import UserTopTabs from "@/features/user/components/ui/userTopTabs";
import SettingsTopTabs from "@/features/user/components/ui/settingsTopTabs";
import OrderDetailsTopNav from "@/features/user/components/ui/orderDetailsTopNav";
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

  const isSettings = pathname.startsWith("/user/settings");
  // Check if inside order details sub-page: /user/orders/[id]
  const isOrderDetails = /^\/user\/orders\/[^/]+$/.test(pathname);

  const centerContent = isSettings ? (
    <SettingsTopTabs />
  ) : isOrderDetails ? (
    <OrderDetailsTopNav />
  ) : (
    <UserTopTabs />
  );

  return (
    <div className="flex flex-col min-h-screen bg-background text-foreground">
      <Header
        centerContent={centerContent}
        actionsContext={isSettings ? "settings" : "default"}
      />
      <div className="w-full px-3 sm:px-6 lg:px-8 py-5 sm:py-8 pb-24 md:pb-12 flex-1">
        <main className="w-full">{children}</main>
      </div>
      {/* Fixed bottom nav — mobile only */}
      <MobileNav />
      <Footer />
    </div>
  );
}

