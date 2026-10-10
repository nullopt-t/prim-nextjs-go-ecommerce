"use client";

import { useTranslations } from "next-intl";
import SettingsShell from "@/features/settings/components/layout/settingsShell";
import Preferences from "@/features/settings/components/ui/preferences";

export default function PreferencesSettingsPage() {
  const t = useTranslations("settings");

  return (
    <SettingsShell activeId="preferences" isSubPage={true}>
      <div className="bg-card border border-border rounded-xl shadow-2xs overflow-hidden">
        <div className="px-4.5 py-3.5 border-b border-border bg-secondary/20">
          <h2 className="font-semibold text-sm text-foreground">
            {t("preferences.title")}
          </h2>
          <p className="text-xs text-muted-foreground mt-0.5">
            {t("preferences.subtitle")}
          </p>
        </div>
        <div className="p-4 sm:p-5">
          <Preferences />
        </div>
      </div>
    </SettingsShell>
  );
}
