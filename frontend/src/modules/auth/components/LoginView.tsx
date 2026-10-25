"use client";

import { useState, useEffect } from "react";
import { LoginEmailForm } from "@/modules/auth/components/ui/loginEmailForm";
import { useRouter } from "@/shared/i18n/navigation";
import { useAuthContext } from "@/modules/auth/model/authContext";
import { useCart } from "@/modules/cart/model/useCart";
import { GuestCartPreview } from "./ui/GuestCartPreview";
import { notify } from "@/shared/lib/notify";
import { useTranslations } from "next-intl";

export function LoginView() {
  const router = useRouter();
  const { startChallenge } = useAuthContext();
  const t = useTranslations("auth");
  const { cartItems } = useCart();
  const [isLoading, setIsLoading] = useState(false);
  const [mergeCart, setMergeCart] = useState(true);

  // Sync merge preference with sessionStorage
  useEffect(() => {
    if (typeof window !== "undefined") {
      const stored = sessionStorage.getItem("merge_cart");
      if (stored !== null) {
        setMergeCart(stored === "true");
      }
    }
  }, []);

  const handleToggleMerge = (val: boolean) => {
    setMergeCart(val);
    if (typeof window !== "undefined") {
      sessionStorage.setItem("merge_cart", String(val));
    }
  };

  const handleEmailSubmit = async (email: string) => {
    if (typeof window !== "undefined") {
      sessionStorage.setItem("identifier", email.trim());
      sessionStorage.setItem("merge_cart", String(mergeCart));
    }

    try {
      setIsLoading(true);
      const res = await startChallenge({ email: email.trim() });
      const searchParams = new URLSearchParams(window.location.search);
      const redirect = searchParams.get("redirect");
      const targetUrl = redirect ? `/auth/verify?redirect=${encodeURIComponent(redirect)}` : "/auth/verify";

      if (res.success) {
        notify.success(t("sign.codeSent"));
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
        <LoginEmailForm
          onSubmit={handleEmailSubmit}
          isLoading={isLoading}
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

