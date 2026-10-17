"use client";

import { Megaphone } from "lucide-react";
import { usePublicSettings } from "@/context/PublicSettingsContext";

/**
 * Storefront announcement ribbon driven by the admin settings
 * (`announcementEnabled` / `announcementText`). Renders nothing when disabled.
 */
export function AnnouncementBanner() {
  const { settings } = usePublicSettings();

  if (!settings?.announcementEnabled || !settings.announcementText?.trim()) {
    return null;
  }

  return (
    <div className="w-full bg-accent-brand text-white">
      <div className="w-full px-3 sm:px-6 lg:px-8 py-2 flex items-center justify-center gap-2 text-center">
        <Megaphone className="size-3.5 shrink-0" />
        <span className="text-[11px] sm:text-xs font-semibold leading-snug">
          {settings.announcementText}
        </span>
      </div>
    </div>
  );
}
