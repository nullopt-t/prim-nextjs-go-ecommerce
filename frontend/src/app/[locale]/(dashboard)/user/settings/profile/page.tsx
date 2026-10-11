"use client";

import { useTranslations } from "next-intl";
import SettingsShell from "@/features/settings/components/layout/settingsShell";
import PersonalInfoForm from "@/features/settings/components/ui/personalInfoForm";

export default function ProfileSettingsPage() {
  return (
    <SettingsShell activeId="profile" isSubPage={true}>
      <div className="bg-card border border-border rounded-xl shadow-2xs overflow-hidden p-4 sm:p-6">
        <PersonalInfoForm />
      </div>
    </SettingsShell>
  );
}
