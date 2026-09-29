"use client";

import DangerZone from "@/features/settings/components/ui/danger";
import PersonalInfoForm from "@/features/settings/components/ui/personalInfoForm";
import Preferences from "@/features/settings/components/ui/preferences";

interface SectionCardProps {
  title: string;
  subtitle?: string;
  children: React.ReactNode;
  titleClassName?: string;
}

function SectionCard({ title, subtitle, children, titleClassName }: SectionCardProps) {
  return (
    <div className="bg-card border border-border rounded-lg shadow-sm overflow-hidden">
      <div className="px-6 py-4 border-b border-border bg-secondary/30">
        <h2 className={`font-semibold text-sm text-foreground ${titleClassName ?? ""}`}>
          {title}
        </h2>
        {subtitle && (
          <p className="text-xs text-muted-foreground mt-0.5">{subtitle}</p>
        )}
      </div>
      <div className="px-6 py-5">{children}</div>
    </div>
  );
}

export default function SettingsLayout() {
  return (
    <div className="w-full flex flex-col gap-6">
      {/* Page header */}
      <div>
        <h1 className="text-xl font-semibold text-foreground">Account Settings</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Manage your personal information and preferences.
        </p>
      </div>

      <SectionCard
        title="Personal Information"
        subtitle="Update your display name and contact details."
      >
        <PersonalInfoForm />
      </SectionCard>

      <SectionCard
        title="Preferences"
        subtitle="Customize your language, theme, and notification settings."
      >
        <Preferences />
      </SectionCard>

      <SectionCard
        title="Danger Zone"
        subtitle="Irreversible account actions."
        titleClassName="text-destructive"
      >
        <DangerZone />
      </SectionCard>
    </div>
  );
}
