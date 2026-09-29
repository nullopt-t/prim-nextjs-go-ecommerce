"use client";

import { useState, useEffect } from "react";
import Form from "@/features/auth/components/ui/form";
import { useRouter } from "@/i18n/navigation";
import { useAuthContext } from "@/context/AuthContext";
import { validateOtpCode } from "../utils/validation";
import { toast } from "sonner";

export function VerifyView() {
  const router = useRouter();
  const { verifyChallenge, resendChallenge } = useAuthContext();
  const [isLoading, setIsLoading] = useState(false);
  const [email, setEmail] = useState<string>("");

  useEffect(() => {
    if (typeof window !== "undefined") {
      const stored = sessionStorage.getItem("identifier");
      if (stored) {
        setEmail(stored);
      }
    }
  }, []);

  const handleSubmit = async (_type: string, payload: { code: string }) => {
    const { isValid, error } = validateOtpCode(payload.code);
    if (!isValid && error) {
      toast.error(error);
      return;
    }

    try {
      setIsLoading(true);
      const res = await verifyChallenge({
        email: email || undefined,
        code: payload.code.trim(),
      });
      if (res.success) {
        toast.success("Welcome back!");
        router.push("/user/overview");
      } else {
        toast.error(res.error || "Invalid verification code");
      }
    } catch {
      toast.error("Failed to verify code. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  const handleResendOtp = async () => {
    if (!email) {
      toast.error("No email specified. Please return to login.");
      return;
    }
    try {
      const res = await resendChallenge({ email });
      if (res.success) {
        toast.success("New verification code sent!");
      } else {
        toast.error(res.error || "Failed to resend code");
      }
    } catch {
      toast.error("Failed to resend verification code");
    }
  };

  const handleEditEmail = () => {
    router.push("/auth");
  };

  return (
    <Form
      formType="verify"
      handleSubmit={handleSubmit}
      isLoading={isLoading}
      userEmail={email}
      onResendOtp={handleResendOtp}
      onEditEmail={handleEditEmail}
    />
  );
}
