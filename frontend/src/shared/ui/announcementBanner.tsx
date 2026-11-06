"use client";

import { useState } from "react";
import { Megaphone, X } from "lucide-react";
import { useTranslations } from "next-intl";
import { usePublicSettings } from "@/modules/settings/model/publicSettingsContext";

/**
 * Storefront announcement ribbon driven by the admin settings
 * (`announcementEnabled` / `announcementText`). Renders nothing when disabled.
 */
export function AnnouncementBanner() {
  const { settings } = usePublicSettings();
  const t = useTranslations("common");
  const [dismissed, setDismissed] = useState(false);

  if (dismissed || !settings?.announcementEnabled || !settings.announcementText?.trim()) {
    return null;
  }

  return (
    <div role="status" aria-live="polite" className="w-full bg-accent-brand text-white">
      <div className="w-full px-3 sm:px-6 lg:px-8 py-2 flex items-center justify-center gap-2 text-center">
        <Megaphone aria-hidden className="size-3.5 shrink-0" />
        <span className="text-[11px] sm:text-xs font-semibold leading-snug">
          {settings.announcementText}
        </span>
        <button
          type="button"
          onClick={() => setDismissed(true)}
          aria-label={t("header.dismiss")}
          className="shrink-0 rounded p-1 transition-colors hover:bg-white/15"
        >
          <X className="size-3.5" aria-hidden />
        </button>
      </div>
    </div>
  );
}
