"use client";

import { useTranslations, useLocale } from "next-intl";
import { Link } from "@/i18n/navigation";
import { ArrowLeft, ArrowRight } from "lucide-react";

interface FormTitleProps {
  type: string;
}

export default function FormTitle({ type }: FormTitleProps) {
  const t = useTranslations("auth");
  const locale = useLocale();
  const isRTL = locale === "ar";

  const title =
    type === "login"
      ? t("sign.title")
      : type === "verify"
      ? t("verify.title")
      : "";

  return (
    <div className="flex items-center gap-2 pb-1">
      {type === "verify" && (
        <Link
          href="/auth"
          aria-label="Back to login"
          className="p-1 -ms-1 text-muted-foreground hover:text-foreground hover:bg-secondary rounded-md transition-colors inline-flex items-center justify-center cursor-pointer"
        >
          {isRTL ? (
            <ArrowRight className="size-4.5" />
          ) : (
            <ArrowLeft className="size-4.5" />
          )}
        </Link>
      )}
      <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-foreground">
        {title}
      </h1>
    </div>
  );
}
