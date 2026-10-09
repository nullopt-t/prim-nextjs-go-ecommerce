"use client";

import { useState } from "react";
import { Toggle } from "@/components/ui/Toggle";
import { useTheme } from "@/context/theme";
import { useRouter, usePathname } from "@/i18n/navigation";
import { useLocale, useTranslations } from "next-intl";

interface PreferenceRowProps {
  label: string;
  description?: string;
  right: React.ReactNode;
}

function PreferenceRow({ label, description, right }: PreferenceRowProps) {
  return (
    <div className="flex items-center justify-between py-3.5 border-b border-border/70 last:border-0 gap-4">
      <div className="min-w-0">
        <p className="text-xs sm:text-sm font-medium text-foreground">{label}</p>
        {description && (
          <p className="text-[11px] sm:text-xs text-muted-foreground mt-0.5">{description}</p>
        )}
      </div>
      <div className="shrink-0">{right}</div>
    </div>
  );
}

export default function Preferences() {
  const { theme, toggle } = useTheme();
  const [emailNotification, setEmailNotification] = useState(false);
  const locale = useLocale();
  const router = useRouter();
  const pathname = usePathname();
  const t = useTranslations("settings.preferences");

  const toggleLang = () => {
    const nextLocale = locale === "en" ? "ar" : "en";
    router.replace(pathname, { locale: nextLocale });
  };

  return (
    <div className="flex flex-col max-w-2xl">
      {/* Language */}
      <PreferenceRow
        label={t("language")}
        description={t("languageDesc")}
        right={
          <button
            type="button"
            onClick={toggleLang}
            className="flex items-center overflow-hidden rounded-full border border-border text-xs font-semibold h-7 shadow-2xs"
          >
            <span
              className={`px-3 py-1 transition-colors ${
                locale === "en"
                  ? "bg-accent-brand text-white"
                  : "bg-transparent text-muted-foreground hover:text-foreground"
              }`}
            >
              EN
            </span>
            <span
              className={`px-3 py-1 transition-colors ${
                locale === "ar"
                  ? "bg-accent-brand text-white"
                  : "bg-transparent text-muted-foreground hover:text-foreground"
              }`}
            >
              عربي
            </span>
          </button>
        }
      />

      {/* Dark mode */}
      <PreferenceRow
        label={t("darkMode")}
        description={t("darkModeDesc")}
        right={
          <Toggle
            isEnabled={theme === "dark"}
            onChange={toggle}
          />
        }
      />

      {/* Email notifications */}
      <PreferenceRow
        label={t("notifications")}
        description={t("notificationsDesc")}
        right={
          <Toggle
            isEnabled={emailNotification}
            onChange={() => setEmailNotification(!emailNotification)}
          />
        }
      />
    </div>
  );
}
