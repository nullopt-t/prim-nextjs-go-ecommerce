"use client";

import { useTranslations } from "next-intl";
import SettingsShell, { SettingsLinksMenu } from "@/features/settings/components/layout/settingsShell";
import PersonalInfoForm from "@/features/settings/components/ui/personalInfoForm";

export default function SettingsIndexPage() {
  const t = useTranslations("settings");

  return (
    <SettingsShell activeId="profile">
      {/* ── Mobile View: Amazon-style clean list of links to sub-pages ── */}
      <div className="block sm:hidden">
        <SettingsLinksMenu />
      </div>

      {/* ── Desktop View: Direct Profile Content with the Top Sub-Tabs ── */}
      <div className="hidden sm:block bg-card border border-border rounded-xl shadow-2xs overflow-hidden p-4 sm:p-6">
        <PersonalInfoForm />
      </div>
    </SettingsShell>
  );
}
