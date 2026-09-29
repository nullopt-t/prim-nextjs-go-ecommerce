"use client";

import { useTranslations } from "next-intl";
import { Link } from "@/i18n/navigation";
import { House, Moon, Sun, ArrowLeft, ArrowRight } from "lucide-react";
import { useTheme } from "@/context/theme";
import { useLocale } from "next-intl";

interface FormTitleProps {
  type: string;
}

export default function FormTitle({ type }: FormTitleProps) {
  const { theme, toggle } = useTheme();
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
    <div className="flex items-center justify-between gap-4 pb-1">
      <div className="flex items-center gap-2">
        {type === "verify" && (
          <Link
            href="/auth"
            aria-label="Back to login"
            className="p-1.5 -ms-1.5 text-muted-foreground hover:text-foreground hover:bg-secondary rounded-lg transition-colors inline-flex items-center justify-center"
          >
            {isRTL ? (
              <ArrowRight className="size-5" />
            ) : (
              <ArrowLeft className="size-5" />
            )}
          </Link>
        )}
        <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-foreground">
          {title}
        </h1>
      </div>

      <div className="flex items-center gap-1.5 text-muted-foreground">
        <Link
          href="/"
          aria-label="Home"
          className="p-2 hover:text-foreground hover:bg-secondary rounded-lg transition-colors inline-flex items-center justify-center"
        >
          <House className="size-4.5" />
        </Link>
        <button
          type="button"
          onClick={toggle}
          aria-label="Toggle theme"
          className="p-2 hover:text-foreground hover:bg-secondary rounded-lg transition-colors cursor-pointer inline-flex items-center justify-center"
        >
          {theme === "light" ? (
            <Sun className="size-4.5" />
          ) : (
            <Moon className="size-4.5" />
          )}
        </button>
      </div>
    </div>
  );
}
