"use client";

import { Fragment, useState, useEffect } from "react";
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

  // Timer countdown for OTP resend
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
    <div className="w-full">
      <div className="mb-2">
        <FormTitle type={formType} />
      </div>
      <div className="mb-6">
        <FormSubtitle
          type={formType}
          email={formType === "verify" ? userEmail || inputs.email : undefined}
          onEditEmail={onEditEmail}
        />
      </div>

      <form
        className="flex flex-col gap-5"
        onSubmit={(e) => {
          e.preventDefault();
          handleSubmit(formType, inputs);
        }}
      >
        {/* Email input for login / register */}
        {(formType === "login" || formType === "register") && (
          <label className="block">
            <span className="block font-medium text-foreground mb-2 text-sm">
              {t("sign.emailLabel")}
            </span>
            <FormInput
              inputObj={{
                type: "email",
                value: inputs.email,
                placeholder: "sign.emailPlaceholder",
                setValue: handleEmail,
                disabled: isLoading,
              }}
            />
          </label>
        )}

        {/* Segmented OTP input for verify */}
        {formType === "verify" && (
          <div className="flex flex-col gap-3">
            <label className="block">
              <span className="block font-medium text-foreground mb-2 text-sm">
                {t("verify.codeLabel")}
              </span>
              <OtpInput
                value={inputs.code}
                onChange={handleCode}
                length={6}
                disabled={isLoading}
                autoFocus
              />
            </label>

            {/* Resend actions */}
            <div className="flex items-center justify-between text-xs sm:text-sm pt-1 text-muted-foreground">
              {canResend ? (
                <button
                  type="button"
                  onClick={handleResend}
                  disabled={isResending || isLoading}
                  className="inline-flex items-center gap-1.5 text-accent-brand hover:underline font-medium cursor-pointer disabled:opacity-50"
                >
                  <RotateCw className={`size-3.5 ${isResending ? "animate-spin" : ""}`} />
                  <span>{t("verify.resend")}</span>
                </button>
              ) : (
                <span className="text-muted-foreground/75">
                  {t("verify.resendIn", { seconds: resendCooldown })}
                </span>
              )}
            </div>
          </div>
        )}

        <div className="mt-2">
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
