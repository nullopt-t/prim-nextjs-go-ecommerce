import { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { SettingsShell, ActiveSessions } from "@/features/settings";
import { ShieldCheck } from "lucide-react";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("settings.sessions");
  return {
    title: `${t("title")} | PRIM`,
  };
}

export default async function SettingsSessionsPage() {
  return (
    <SettingsShell activeId="sessions" isSubPage={true}>
      <div className="rounded-2xl border border-border/80 bg-card shadow-2xs overflow-hidden p-4 sm:p-6">
        <ActiveSessions />
      </div>
    </SettingsShell>
  );
}
