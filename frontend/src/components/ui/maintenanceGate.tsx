"use client";

import { ShieldAlert } from "lucide-react";
import { usePublicSettings } from "@/context/PublicSettingsContext";

/**
 * Gates the storefront behind the admin-configured maintenance mode. While
 * maintenance mode is enabled, visitors see the maintenance message instead of
 * the shop. Admin, auth, and other non-shop areas are unaffected because this
 * wrapper is only mounted around the shop layout.
 */
export function MaintenanceGate({ children }: { children: React.ReactNode }) {
  const { settings, loading } = usePublicSettings();

  if (!loading && settings?.maintenanceMode) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background text-foreground px-6">
        <div className="max-w-md w-full text-center space-y-5">
          <div className="mx-auto size-14 rounded-2xl bg-amber-500/10 border border-amber-500/20 flex items-center justify-center">
            <ShieldAlert className="size-7 text-amber-500" />
          </div>
          <div className="space-y-2">
            <h1 className="text-xl font-bold tracking-tight">
              {settings.storeName || "Store"} — Maintenance
            </h1>
            <p className="text-sm text-muted-foreground leading-relaxed">
              {settings.maintenanceMessage ||
                "We are currently performing scheduled maintenance. Please check back shortly."}
            </p>
          </div>
        </div>
      </div>
    );
  }

  return <>{children}</>;
}
