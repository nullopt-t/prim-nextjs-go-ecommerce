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
  const t = await getTranslations("settings.sessions");

  return (
    <SettingsShell activeId="sessions" isSubPage={true}>
      <div className="rounded-2xl border border-border/80 bg-card shadow-2xs overflow-hidden">
        {/* Card Header */}
        <div className="px-4.5 py-3.5 border-b border-border/70 bg-secondary/15 flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <div className="p-1.5 rounded-lg bg-secondary text-foreground">
              <ShieldCheck className="size-4" />
            </div>
            <div>
              <h2 className="text-sm font-semibold text-foreground tracking-tight">{t("title")}</h2>
              <p className="text-[11px] text-muted-foreground">{t("subtitle")}</p>
            </div>
          </div>
        </div>

        {/* Card Body */}
        <div className="p-4 sm:p-5">
          <ActiveSessions />
        </div>
      </div>
    </SettingsShell>
  );
}
