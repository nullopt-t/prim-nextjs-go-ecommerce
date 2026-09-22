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
      className="bg-primary text-primary-foreground rounded-md py-3 font-medium capitalize hover:opacity-90 transition-opacity cursor-pointer w-full"
      onClick={(e) => {
        e.preventDefault();
        handle(type, payload);
      }}
    >
      {buttonTxt}
    </button>
  );
}
