"use client";

import Form from "@/features/auth/components/ui/form";
import { useRouter } from "@/i18n/navigation";

export function VerifyView() {
  const router = useRouter();

  const validateCode = (code: string) => {
    if (!code.trim()) {
      return "Verification code is required.";
    }
    const codeRegex = /^\d{6}$/;
    if (!codeRegex.test(code)) {
      return "Verification code must contain 6 digits.";
    }
    return null;
  };

  const handleSubmit = async (type: string, payload: { code: string }) => {
    const codeError = validateCode(payload.code);
    if (codeError) {
      alert(codeError);
      return;
    }

    router.push("/user/overview");
  };

  return <Form formType="verify" handleSubmit={handleSubmit} />;
}
