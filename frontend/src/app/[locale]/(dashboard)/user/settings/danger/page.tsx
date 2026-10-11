"use client";

import { useTranslations } from "next-intl";
import SettingsShell from "@/features/settings/components/layout/settingsShell";
import DangerZone from "@/features/settings/components/ui/danger";

export default function DangerSettingsPage() {
  return (
    <SettingsShell activeId="danger" isSubPage={true}>
      <div className="bg-card border border-destructive/30 rounded-xl shadow-2xs overflow-hidden p-4 sm:p-6">
        <DangerZone />
      </div>
    </SettingsShell>
  );
}
