"use client";

import { useTranslations } from "next-intl";
import { Loader2 } from "lucide-react";

interface FormButtonProps {
  type: string;
  isLoading?: boolean;
  disabled?: boolean;
  label?: string;
}

export default function FormButton({
  type,
  isLoading = false,
  disabled = false,
  label,
}: FormButtonProps) {
  const t = useTranslations("auth");
  const buttonTxt =
    label ??
    (type === "login"
      ? t("sign.signButton")
      : type === "verify"
      ? t("verify.verifyButton")
      : "");

  return (
    <button
      type="submit"
      disabled={isLoading || disabled}
      className="w-full h-11 rounded-md bg-primary text-primary-foreground font-medium text-sm transition-colors hover:bg-primary/90 flex items-center justify-center gap-2 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
    >
      {isLoading && <Loader2 className="size-4 animate-spin" />}
      <span>{buttonTxt}</span>
    </button>
  );
}
