"use client";

import Form from "@/features/auth/components/ui/form";
import { useRouter } from "@/i18n/navigation";
import { useAuthContext } from "@/context/AuthContext";
import { validateOtpCode } from "../domain/authDomain";
import { toast } from "sonner";

export function VerifyView() {
  const router = useRouter();
  const { verifyChallenge } = useAuthContext();

  const handleSubmit = async (_type: string, payload: { code: string }) => {
    const { isValid, error } = validateOtpCode(payload.code);
    if (!isValid && error) {
      toast.error(error);
      return;
    }

    const email = typeof window !== "undefined" ? sessionStorage.getItem("identifier") || undefined : undefined;

    try {
      const res = await verifyChallenge({ email, code: payload.code });
      if (res.success) {
        toast.success("Welcome back!");
        router.push("/user/overview");
      } else {
        toast.error(res.error || "Invalid verification code");
      }
    } catch {
      toast.error("Failed to verify code. Please try again.");
    }
  };

  return <Form formType="verify" handleSubmit={handleSubmit} />;
}
