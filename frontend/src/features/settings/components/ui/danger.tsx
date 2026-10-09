"use client";

import { useAuthContext } from "@/context/AuthContext";
import { useRouter } from "@/i18n/navigation";
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
  const { logout } = useAuthContext();
  const router = useRouter();
  const t = useTranslations("settings.danger");

  const handleLogoutAll = async () => {
    await logout();
    router.push("/auth");
  };

  return (
    <div className="flex flex-col max-w-2xl">
      <DangerRow
        title={t("signOutAll")}
        description={t("signOutAllDesc")}
        action={
          <button
            type="button"
            onClick={handleLogoutAll}
            className="px-3.5 py-1.5 rounded-lg text-xs font-semibold border border-destructive/40 text-destructive hover:bg-destructive/10 transition-colors shadow-2xs"
          >
            {t("signOutAllButton")}
          </button>
        }
      />
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
