"use client";

import { useTranslations } from "next-intl";
import SettingsShell from "@/features/settings/components/layout/settingsShell";
import Preferences from "@/features/settings/components/ui/preferences";

export default function PreferencesSettingsPage() {
  return (
    <SettingsShell activeId="preferences" isSubPage={true}>
      <div className="bg-card border border-border rounded-xl shadow-2xs overflow-hidden p-4 sm:p-6">
        <Preferences />
      </div>
    </SettingsShell>
  );
}
