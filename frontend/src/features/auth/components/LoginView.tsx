"use client";

import Form from "@/features/auth/components/ui/form";
import { useRouter } from "@/i18n/navigation";

export function LoginView() {
  const router = useRouter();

  const validateEmail = (email: string) => {
    if (!email.trim()) {
      return "Email is required.";
    }
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      return "Please enter a valid email address.";
    }
    return null;
  };

  const handleSubmit = async (type: string, payload: { email: string }) => {
    const emailError = validateEmail(payload.email);
    if (emailError) {
      alert(emailError);
      return;
    }

    if (typeof window !== "undefined") {
      sessionStorage.setItem("identifier", payload.email);
    }
    router.push("/auth/verify");
  };

  return <Form formType="login" handleSubmit={handleSubmit} />;
}
