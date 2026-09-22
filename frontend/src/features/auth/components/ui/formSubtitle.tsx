"use client";

import { useTranslations } from "next-intl";

export default function FormSubtitle({ type }: { type: string }) {
  const t = useTranslations("auth");

  const subTitle =
    type === "login"
      ? t("sign.subtitle")
      : type === "verify"
      ? t("sign.subtitle")
      : "";

  return (
    <p className="text-muted-foreground text-txt-sm md:text-txt-md lg:text-txt-lg">
      {subTitle}
    </p>
  );
}
