"use client";

import { useTranslations } from "next-intl";
import SettingsShell from "@/features/settings/components/layout/settingsShell";
import PersonalInfoForm from "@/features/settings/components/ui/personalInfoForm";

export default function ProfileSettingsPage() {
  const t = useTranslations("settings");

  return (
    <SettingsShell activeId="profile" isSubPage={true}>
      <div className="bg-card border border-border rounded-xl shadow-2xs overflow-hidden">
        <div className="px-4.5 py-3.5 border-b border-border bg-secondary/20">
          <h2 className="font-semibold text-sm text-foreground">
            {t("personalInfo.title")}
          </h2>
          <p className="text-xs text-muted-foreground mt-0.5">
            {t("personalInfo.subtitle")}
          </p>
        </div>
        <div className="p-4 sm:p-5">
          <PersonalInfoForm />
        </div>
      </div>
    </SettingsShell>
  );
}
