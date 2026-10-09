"use client";

import { useTranslations } from "next-intl";
import PersonalInfoForm from "@/features/settings/components/ui/personalInfoForm";
import Preferences from "@/features/settings/components/ui/preferences";
import DangerZone from "@/features/settings/components/ui/danger";

interface SectionCardProps {
  title: string;
  subtitle?: string;
  children: React.ReactNode;
  titleClassName?: string;
}

function SectionCard({ title, subtitle, children, titleClassName }: SectionCardProps) {
  return (
    <div className="bg-card border border-border rounded-xl shadow-2xs overflow-hidden">
      <div className="px-5 py-4 border-b border-border bg-secondary/20">
        <h2 className={`font-semibold text-sm text-foreground ${titleClassName ?? ""}`}>
          {title}
        </h2>
        {subtitle && (
          <p className="text-xs text-muted-foreground mt-0.5">{subtitle}</p>
        )}
      </div>
      <div className="p-5 sm:p-6">{children}</div>
    </div>
  );
}

export default function SettingsLayout() {
  const t = useTranslations("settings");

  return (
    <div className="w-full flex flex-col gap-6 pb-12">
      {/* ── Page Header Strip ── */}
      <div className="pb-2 border-b border-border/60">
        <h1 className="text-2xl font-bold tracking-tight text-foreground">{t("title")}</h1>
        <p className="text-sm text-muted-foreground mt-0.5">
          {t("subtitle")}
        </p>
      </div>

      {/* ── Personal Info ── */}
      <SectionCard
        title={t("personalInfo.title")}
        subtitle={t("personalInfo.subtitle")}
      >
        <PersonalInfoForm />
      </SectionCard>

      {/* ── Preferences ── */}
      <SectionCard
        title={t("preferences.title")}
        subtitle={t("preferences.subtitle")}
      >
        <Preferences />
      </SectionCard>

      {/* ── Danger Zone ── */}
      <SectionCard
        title={t("danger.title")}
        subtitle={t("danger.subtitle")}
        titleClassName="text-destructive font-bold"
      >
        <DangerZone />
      </SectionCard>
    </div>
  );
}
