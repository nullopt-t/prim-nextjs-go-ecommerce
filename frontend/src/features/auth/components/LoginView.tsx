"use client";

import { useState } from "react";
import Form from "@/features/auth/components/ui/form";
import { useRouter } from "@/i18n/navigation";
import { useAuthContext } from "@/context/AuthContext";
import { validateEmail } from "../utils/validation";
import { toast } from "sonner";

export function LoginView() {
  const router = useRouter();
  const { startChallenge } = useAuthContext();
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (_type: string, payload: { email: string }) => {
    const { isValid, error } = validateEmail(payload.email);
    if (!isValid && error) {
      toast.error(error);
      return;
    }

    if (typeof window !== "undefined") {
      sessionStorage.setItem("identifier", payload.email.trim());
    }

    try {
      setIsLoading(true);
      const res = await startChallenge({ email: payload.email.trim() });
      const searchParams = new URLSearchParams(window.location.search);
      const redirect = searchParams.get("redirect");
      const targetUrl = redirect ? `/auth/verify?redirect=${encodeURIComponent(redirect)}` : "/auth/verify";

      if (res.success) {
        toast.success("Verification code sent to your email!");
        router.push(targetUrl);
      } else {
        // Fallback navigation if offline or demo mode
        router.push(targetUrl);
      }
    } catch {
      const searchParams = new URLSearchParams(window.location.search);
      const redirect = searchParams.get("redirect");
      const targetUrl = redirect ? `/auth/verify?redirect=${encodeURIComponent(redirect)}` : "/auth/verify";
      router.push(targetUrl);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Form
      formType="login"
      handleSubmit={handleSubmit}
      isLoading={isLoading}
    />
  );
}
