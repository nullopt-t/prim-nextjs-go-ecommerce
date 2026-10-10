"use client";

import { useAuthContext } from "@/context/AuthContext";
import { useTranslations } from "next-intl";

export default function PersonalInfoForm() {
  const { user } = useAuthContext();
  const t = useTranslations("settings.personalInfo");

  return (
    <div className="flex flex-col gap-8 w-full">
      <form className="flex flex-col gap-5" onSubmit={(e) => e.preventDefault()}>
        {/* Notice */}
        <div className="rounded-xl border border-border/70 bg-secondary/30 px-4 py-3 text-xs text-muted-foreground leading-relaxed">
          {t("notice")}
        </div>

        {/* Display Name */}
        <div className="flex flex-col gap-1.5">
          <label htmlFor="display-name" className="text-xs font-medium text-foreground">
            {t("displayName")}
          </label>
          <input
            id="display-name"
            type="text"
            defaultValue={user?.name || ""}
            readOnly
            disabled
            placeholder={t("displayName")}
            className="h-10 w-full rounded-xl border border-border bg-secondary/30 px-3.5 text-xs sm:text-sm text-foreground placeholder:text-muted-foreground disabled:opacity-75 disabled:cursor-not-allowed focus:outline-none transition"
          />
        </div>

        {/* Email */}
        <div className="flex flex-col gap-1.5">
          <label htmlFor="email" className="text-xs font-medium text-foreground">
            {t("email")}
          </label>
          <input
            id="email"
            type="email"
            value={user?.email || ""}
            readOnly
            disabled
            className="h-10 w-full rounded-xl border border-border bg-secondary/30 px-3.5 text-xs sm:text-sm text-foreground placeholder:text-muted-foreground disabled:opacity-75 disabled:cursor-not-allowed focus:outline-none transition"
          />
          <p className="text-[11px] text-muted-foreground">
            {t("emailNotice")}
          </p>
        </div>

        {/* Phone */}
        <div className="flex flex-col gap-1.5">
          <label htmlFor="phone" className="text-xs font-medium text-foreground">
            {t("phone")}
          </label>
          <input
            id="phone"
            type="tel"
            value={user?.phone || ""}
            readOnly
            disabled
            placeholder={t("notSet")}
            className="h-10 w-full rounded-xl border border-border bg-secondary/30 px-3.5 text-xs sm:text-sm text-foreground placeholder:text-muted-foreground disabled:opacity-75 disabled:cursor-not-allowed focus:outline-none transition"
          />
          <p className="text-[11px] text-muted-foreground">
            {t("phoneNotice")}
          </p>
        </div>
      </form>
    </div>
  );
}
