"use client";

import { useTranslations } from "next-intl";

interface FormButtonProps {
  type: string;
  payload: any;
  handle: (type: string, payload: any) => void;
}

export default function FormButton({ type, payload, handle }: FormButtonProps) {
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
      className="bg-primary text-primary-foreground rounded-xl py-3.5 font-semibold text-sm capitalize hover:bg-primary/90 transition-all shadow-sm cursor-pointer w-full"
      onClick={(e) => {
        e.preventDefault();
        handle(type, payload);
      }}
    >
      {buttonTxt}
    </button>
  );
}
