"use client";

import { useState, useEffect } from "react";
import FormButton from "@/features/auth/components/ui/formButton";
import FormTitle from "@/features/auth/components/ui/formTitle";
import FormSubtitle from "@/features/auth/components/ui/formSubtitle";
import FormInput from "@/features/auth/components/ui/formInput";
import OtpInput from "@/features/auth/components/ui/OtpInput";
import { useTranslations } from "next-intl";
import { AuthFormType } from "@/features/auth/types";
import { RotateCw } from "lucide-react";

interface FormProps {
  formType: AuthFormType;
  handleSubmit: (type: string, payload: { email: string; code: string }) => Promise<void> | void;
  isLoading?: boolean;
  userEmail?: string;
  onResendOtp?: () => Promise<void> | void;
  onEditEmail?: () => void;
}

export default function Form({
  formType,
  handleSubmit,
  isLoading = false,
  userEmail,
  onResendOtp,
  onEditEmail,
}: FormProps) {
  const t = useTranslations("auth");

  const [inputs, setInputs] = useState({
    email: userEmail || "",
    code: "",
  });

  const [resendCooldown, setResendCooldown] = useState(30);
  const [canResend, setCanResend] = useState(false);
  const [isResending, setIsResending] = useState(false);

  useEffect(() => {
    if (formType !== "verify") return;
    if (resendCooldown <= 0) {
      setCanResend(true);
      return;
    }

    setCanResend(false);
    const interval = setInterval(() => {
      setResendCooldown((prev) => prev - 1);
    }, 1000);

    return () => clearInterval(interval);
  }, [formType, resendCooldown]);

  const handleEmail = (emailValue: string) => {
    setInputs((prev) => ({ ...prev, email: emailValue }));
  };

  const handleCode = (codeValue: string) => {
    setInputs((prev) => ({ ...prev, code: codeValue }));
  };

  const handleResend = async () => {
    if (!canResend || isResending || !onResendOtp) return;
    try {
      setIsResending(true);
      await onResendOtp();
      setResendCooldown(45);
      setCanResend(false);
    } finally {
      setIsResending(false);
    }
  };

  return (
    <div className="w-full flex flex-col gap-5">
      <div className="flex flex-col gap-1">
        <FormTitle type={formType} />
        <FormSubtitle
          type={formType}
          email={formType === "verify" ? userEmail || inputs.email : undefined}
          onEditEmail={onEditEmail}
        />
      </div>

      <form
        className="flex flex-col gap-4"
        onSubmit={(e) => {
          e.preventDefault();
          handleSubmit(formType, inputs);
        }}
      >
        {/* Email input for login */}
        {(formType === "login" || formType === "register") && (
          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-medium text-foreground">
              {t("sign.emailLabel")}
            </label>
            <FormInput
              inputObj={{
                type: "email",
                value: inputs.email,
                placeholder: "sign.emailPlaceholder",
                setValue: handleEmail,
                disabled: isLoading,
              }}
            />
          </div>
        )}

        {/* OTP input for verify */}
        {formType === "verify" && (
          <div className="flex flex-col gap-2.5">
            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-foreground">
                {t("verify.codeLabel")}
              </label>
              <OtpInput
                value={inputs.code}
                onChange={handleCode}
                length={6}
                disabled={isLoading}
                autoFocus
              />
            </div>

            <div className="flex items-center justify-between text-xs text-muted-foreground pt-1">
              {canResend ? (
                <button
                  type="button"
                  onClick={handleResend}
                  disabled={isResending || isLoading}
                  className="inline-flex items-center gap-1 text-accent-brand hover:underline font-medium cursor-pointer disabled:opacity-50"
                >
                  <RotateCw className={`size-3 ${isResending ? "animate-spin" : ""}`} />
                  <span>{t("verify.resend")}</span>
                </button>
              ) : (
                <span>{t("verify.resendIn", { seconds: resendCooldown })}</span>
              )}
            </div>
          </div>
        )}

        <div className="pt-1">
          <FormButton
            type={formType}
            payload={inputs}
            handle={handleSubmit}
            isLoading={isLoading}
            disabled={
              formType === "verify"
                ? inputs.code.trim().length !== 6
                : !inputs.email.trim()
            }
          />
        </div>
      </form>
    </div>
  );
}
