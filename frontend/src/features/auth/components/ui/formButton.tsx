"use client";

import { useTranslations } from "next-intl";
import { Loader2 } from "lucide-react";

interface FormButtonProps {
  type: string;
  payload: any;
  handle: (type: string, payload: any) => void;
  isLoading?: boolean;
  disabled?: boolean;
}

export default function FormButton({
  type,
  payload,
  handle,
  isLoading = false,
  disabled = false,
}: FormButtonProps) {
  const t = useTranslations("auth");
  const buttonTxt =
    type === "login"
      ? t("sign.signButton")
      : type === "verify"
      ? t("verify.verifyButton")
      : "";

  return (
    <button
      type="submit"
      disabled={isLoading || disabled}
      className="relative flex items-center justify-center gap-2 bg-primary text-primary-foreground rounded-xl py-3.5 px-6 font-semibold text-sm sm:text-base capitalize hover:bg-primary/90 active:scale-[0.99] transition-all shadow-sm cursor-pointer w-full disabled:opacity-60 disabled:cursor-not-allowed disabled:active:scale-100"
      onClick={(e) => {
        e.preventDefault();
        if (!isLoading && !disabled) {
          handle(type, payload);
        }
      }}
    >
      {isLoading && <Loader2 className="size-4.5 animate-spin" />}
      <span>{buttonTxt}</span>
    </button>
  );
}
