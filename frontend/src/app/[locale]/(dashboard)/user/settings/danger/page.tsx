"use client";

import { useTranslations } from "next-intl";
import SettingsShell from "@/features/settings/components/layout/settingsShell";
import DangerZone from "@/features/settings/components/ui/danger";

export default function DangerSettingsPage() {
  const t = useTranslations("settings");

  return (
    <SettingsShell activeId="danger" isSubPage={true}>
      <div className="bg-card border border-destructive/30 rounded-xl shadow-2xs overflow-hidden">
        <div className="px-4.5 py-3.5 border-b border-destructive/20 bg-destructive/5">
          <h2 className="font-bold text-sm text-destructive">
            {t("danger.title")}
          </h2>
          <p className="text-xs text-muted-foreground mt-0.5">
            {t("danger.subtitle")}
          </p>
        </div>
        <div className="p-4 sm:p-5">
          <DangerZone />
        </div>
      </div>
    </SettingsShell>
  );
}
