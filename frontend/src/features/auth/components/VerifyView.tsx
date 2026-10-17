"use client";

import { useState, useEffect } from "react";
import Form from "@/features/auth/components/ui/form";
import { useRouter } from "@/i18n/navigation";
import { useAuthContext } from "@/context/AuthContext";
import { useCart } from "@/hooks/useCart";
import { GuestCartPreview } from "./ui/GuestCartPreview";
import { validateOtpCode } from "../utils/validation";
import { toast } from "sonner";
import { useTranslations } from "next-intl";

export function VerifyView() {
  const router = useRouter();
  const { verifyChallenge, resendChallenge } = useAuthContext();
  const t = useTranslations("auth");
  const { cartItems } = useCart();
  const [isLoading, setIsLoading] = useState(false);
  const [email, setEmail] = useState<string>("");
  const [mergeCart, setMergeCart] = useState(true);

  useEffect(() => {
    if (typeof window !== "undefined") {
      const storedEmail = sessionStorage.getItem("identifier");
      if (storedEmail) {
        setEmail(storedEmail);
      }
      const storedMerge = sessionStorage.getItem("merge_cart");
      if (storedMerge !== null) {
        setMergeCart(storedMerge === "true");
      }
    }
  }, []);

  const handleToggleMerge = (val: boolean) => {
    setMergeCart(val);
    if (typeof window !== "undefined") {
      sessionStorage.setItem("merge_cart", String(val));
    }
  };

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
        mergeCart,
      });
      if (res.success) {
        toast.success(t("verify.welcomeBack"));
        // Clear auth temp storage
        if (typeof window !== "undefined") {
          sessionStorage.removeItem("identifier");
          sessionStorage.removeItem("merge_cart");
        }
        const searchParams = new URLSearchParams(window.location.search);
        const redirectParam = searchParams.get("redirect");
        const destination = redirectParam ? decodeURIComponent(redirectParam) : "/";
        router.push(destination as any);
      } else {
        toast.error(res.error || t("verify.invalidCode"));
      }
    } catch {
      toast.error(t("verify.verifyFailed"));
    } finally {
      setIsLoading(false);
    }
  };

  const handleResendOtp = async () => {
    if (!email) {
      toast.error(t("verify.noEmail"));
      return;
    }
    try {
      const res = await resendChallenge({ email });
      if (res.success) {
        toast.success(t("verify.codeResent"));
      } else {
        toast.error(res.error || t("verify.resendError"));
      }
    } catch {
      toast.error(t("verify.resendFailed"));
    }
  };

  const handleEditEmail = () => {
    router.push("/auth");
  };

  const hasGuestItems = cartItems && cartItems.length > 0;

  return (
    <div
      className={`w-full ${
        hasGuestItems
          ? "max-w-5xl mx-auto grid grid-cols-1 lg:grid-cols-5 gap-8 lg:gap-14 items-start"
          : "max-w-[420px] mx-auto"
      }`}
    >
      <div className={`w-full ${hasGuestItems ? "lg:col-span-2" : ""}`}>
        <Form
          formType="verify"
          handleSubmit={handleSubmit}
          isLoading={isLoading}
          userEmail={email}
          onResendOtp={handleResendOtp}
          onEditEmail={handleEditEmail}
        />
      </div>

      {hasGuestItems && (
        <div className="lg:col-span-3">
          <GuestCartPreview
            mergeCart={mergeCart}
            onToggleMerge={handleToggleMerge}
          />
        </div>
      )}
    </div>
  );
}

