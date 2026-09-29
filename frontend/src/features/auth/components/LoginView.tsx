"use client";

import Form from "@/features/auth/components/ui/form";
import { useRouter } from "@/i18n/navigation";
import { useAuthContext } from "@/context/AuthContext";
import { validateEmail } from "../domain/authDomain";
import { toast } from "sonner";

export function LoginView() {
  const router = useRouter();
  const { startChallenge } = useAuthContext();

  const handleSubmit = async (_type: string, payload: { email: string }) => {
    const { isValid, error } = validateEmail(payload.email);
    if (!isValid && error) {
      toast.error(error);
      return;
    }

    if (typeof window !== "undefined") {
      sessionStorage.setItem("identifier", payload.email);
    }

    try {
      const res = await startChallenge({ email: payload.email });
      if (res.success) {
        toast.success("Verification code sent to your email!");
        router.push("/auth/verify");
      } else {
        // Fallback navigation if offline or demo mode
        router.push("/auth/verify");
      }
    } catch {
      router.push("/auth/verify");
    }
  };

  return <Form formType="login" handleSubmit={handleSubmit} />;
}
