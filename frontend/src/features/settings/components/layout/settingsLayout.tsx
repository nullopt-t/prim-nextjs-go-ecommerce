"use client";

import { Title } from "@/components/ui/title";
import DangerZone from "@/features/settings/components/ui/danger";
import PersonalInfoForm from "@/features/settings/components/ui/personalInfoForm";
import Preferences from "@/features/settings/components/ui/preferences";
import { useTranslations } from "next-intl";

export default function SettingsLayout() {
  const t = useTranslations("settings");

  return (
    <div className="flex flex-col gap-10">
      <Title
        title={t("settings.title")}
        subtitle={t("settings.description")}
      />
      <div>
        <Title title={t("settings.personalInformation")} />
        <PersonalInfoForm />
      </div>
      <div>
        <Title title={t("settings.preferences")} />
        <Preferences />
      </div>
      <div>
        <Title
          title={t("settings.dangerZone")}
          textColor="text-destructive"
        />
        <DangerZone />
      </div>
    </div>
  );
}
