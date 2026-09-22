"use client";

import { Text } from "@/components/ui/text";
import { Toggle } from "@/components/ui/Toggle";
import { useState } from "react";
import { useTranslations, useLocale } from "next-intl";
import { useTheme } from "@/context/theme";
import { useRouter, usePathname } from "@/i18n/navigation";

export default function Preferences() {
  const { theme, toggle } = useTheme();
  const [emailNotification, setEmailNotification] = useState(false);
  const locale = useLocale();
  const router = useRouter();
  const pathname = usePathname();
  const t = useTranslations("settings");

  const toggleLang = () => {
    const nextLocale = locale === "en" ? "ar" : "en";
    router.replace(pathname, { locale: nextLocale });
  };

  return (
    <div className="flex flex-col gap-5 max-w-xl">
      <div className="flex justify-between items-center">
        <Text text={t("settings.language")} />
        <button
          type="button"
          onClick={toggleLang}
          className="cursor-pointer font-medium text-foreground hover:text-accent-brand transition-colors"
        >
          <span className={locale === "en" ? "text-accent-brand font-bold" : ""}>
            En
          </span>
          <span> / </span>
          <span className={locale === "ar" ? "text-accent-brand font-bold" : ""}>
            ع
          </span>
        </button>
      </div>
      <div className="flex justify-between items-center">
        <Text text={t("settings.darkMode")} />
        <Toggle isEnabled={theme === "dark"} onChange={toggle} />
      </div>
      <div className="flex justify-between items-center">
        <Text text={t("settings.emailNotifications")} />
        <Toggle
          isEnabled={emailNotification}
          onChange={(e) => setEmailNotification(e.target.checked)}
        />
      </div>
    </div>
  );
}
