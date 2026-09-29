"use client";

import { useTranslations } from "next-intl";

interface FormSubtitleProps {
  type: string;
  email?: string;
  onEditEmail?: () => void;
}

export default function FormSubtitle({
  type,
  email,
  onEditEmail,
}: FormSubtitleProps) {
  const t = useTranslations("auth");

  if (type === "verify" && email) {
    return (
      <div className="flex flex-col gap-1 text-sm text-muted-foreground leading-relaxed">
        <p>{t("verify.subtitle")}</p>
        <div className="flex items-center gap-2 flex-wrap pt-0.5">
          <span className="font-semibold text-foreground underline underline-offset-2">
            {email}
          </span>
          {onEditEmail && (
            <button
              type="button"
              onClick={onEditEmail}
              className="text-xs text-accent-brand hover:underline font-medium cursor-pointer"
            >
              {t("verify.changeEmail")}
            </button>
          )}
        </div>
      </div>
    );
  }

  const subTitle =
    type === "login"
      ? t("sign.subtitle")
      : type === "verify"
      ? t("verify.subtitle")
      : "";

  return (
    <p className="text-sm sm:text-base text-muted-foreground leading-relaxed">
      {subTitle}
    </p>
  );
}
