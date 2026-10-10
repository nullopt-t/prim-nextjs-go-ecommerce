"use client";

import { useTranslations } from "next-intl";

interface DangerRowProps {
  title: string;
  description: string;
  action: React.ReactNode;
}

function DangerRow({ title, description, action }: DangerRowProps) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 py-4 border-b border-border/70 last:border-0">
      <div className="min-w-0">
        <p className="text-xs sm:text-sm font-semibold text-foreground">{title}</p>
        <p className="text-[11px] sm:text-xs text-muted-foreground mt-0.5 max-w-md leading-relaxed">{description}</p>
      </div>
      <div className="sm:shrink-0">{action}</div>
    </div>
  );
}

export default function DangerZone() {
  const t = useTranslations("settings.danger");

  return (
    <div className="flex flex-col w-full">
      <DangerRow
        title={t("deleteAccount")}
        description={t("deleteAccountDesc")}
        action={
          <button
            type="button"
            disabled
            title={t("deleteNotice")}
            className="px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-destructive/60 text-white opacity-50 cursor-not-allowed shadow-2xs"
          >
            {t("deleteAccountButton")}
          </button>
        }
      />
    </div>
  );
}
