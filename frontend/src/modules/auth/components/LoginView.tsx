"use client";

import { useState, useEffect } from "react";
import { LoginEmailForm } from "@/modules/auth/components/ui/loginEmailForm";
import { useRouter } from "@/shared/i18n/navigation";
import { useAuthContext } from "@/modules/auth/model/authContext";
import { useCart } from "@/modules/cart/model/useCart";
import { GuestCartPreview } from "./ui/GuestCartPreview";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { useTranslations } from "next-intl";
import { Fingerprint, Loader2 } from "lucide-react";
import { authService } from "@/modules/auth/api/authService";
import { isPasskeySupported, signInWithPasskey } from "@/modules/auth/lib/passkey";

export function LoginView() {
  const router = useRouter();
  const { startChallenge } = useAuthContext();
  const t = useTranslations("auth");
  const { cartItems } = useCart();
  const [isLoading, setIsLoading] = useState(false);
  const [mergeCart, setMergeCart] = useState(true);
  const [passkeySupported, setPasskeySupported] = useState(false);
  const [passkeyLoading, setPasskeyLoading] = useState(false);

  useEffect(() => {
    isPasskeySupported().then(setPasskeySupported).catch(() => setPasskeySupported(false));
  }, []);

  // Sync merge preference with sessionStorage on mount
  useEffect(() => {
    if (typeof window !== "undefined") {
      const stored = sessionStorage.getItem("merge_cart");
      if (stored !== null) {
        setMergeCart(stored === "true");
      }
    }
  }, []);

  // Sync sessionStorage when mergeCart state changes to avoid stale data
  useEffect(() => {
    if (typeof window !== "undefined") {
      sessionStorage.setItem("merge_cart", String(mergeCart));
    }
  }, [mergeCart]);

  const handleToggleMerge = (val: boolean) => {
    setMergeCart(val);
    if (typeof window !== "undefined") {
      sessionStorage.setItem("merge_cart", String(val));
    }
  };

  const handleEmailSubmit = async (email: string) => {
    try {
      setIsLoading(true);
      const res = await startChallenge({ email: email.trim() });
      if (!res.success) {
        notify.error(apiErrorMessage(res, t("sign.sendFailed")));
        return;
      }
      // Store identifier in sessionStorage for pre-filling on return
      // Clear stale entries on new submission
      if (typeof window !== "undefined") {
        sessionStorage.setItem("identifier", email.trim());
        sessionStorage.setItem("merge_cart", String(mergeCart));
      }
      const searchParams = new URLSearchParams(window.location.search);
      const redirect = searchParams.get("redirect");
      const targetUrl = redirect ? `/auth/verify?redirect=${encodeURIComponent(redirect)}` : "/auth/verify";
      notify.success(t("sign.codeSent"));
      router.push(targetUrl);
    } catch {
      notify.error(t("sign.sendFailed"));
    } finally {
      setIsLoading(false);
    }
  };

  const handlePasskey = async () => {
    setPasskeyLoading(true);
    try {
      const result = await signInWithPasskey({
        fetchOptions: async () => {
          const res = await authService.getPasskeyAuthenticationOptions();
          return { ...res.data, userVerification: undefined };
        },
        verify: async (credential, requestId) => {
          try {
            await authService.verifyPasskeyAssertion(requestId, credential);
            return { ok: true };
          } catch (err) {
            return { ok: false, error: apiErrorMessage(err, t("passkey.failed")) };
          }
        },
      });
      if (result.ok) {
        const searchParams = new URLSearchParams(window.location.search);
        const redirect = searchParams.get("redirect");
        notify.success(t("passkey.success"));
        router.replace(redirect || "/");
      } else if (result.error === "cancelled") {
        notify.info(t("passkey.cancelled"));
      } else {
        notify.error(result.error || t("passkey.failed"));
      }
    } catch (err) {
      if (err instanceof DOMException && (err.name === "AbortError" || err.name === "NotAllowedError")) {
        notify.info(t("passkey.cancelled"));
      } else {
        notify.error(apiErrorMessage(err, t("passkey.failed")));
      }
    } finally {
      setPasskeyLoading(false);
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

        {passkeySupported && (
          <div className="mt-4">
            <div className="relative my-2 flex items-center gap-3 text-[11px] text-muted-foreground">
              <span className="h-px flex-1 bg-border" />
              {t("passkey.or")}
              <span className="h-px flex-1 bg-border" />
            </div>
            <button
              type="button"
              onClick={() => void handlePasskey()}
              disabled={passkeyLoading}
              className="flex w-full h-11 items-center justify-center gap-2 rounded-md border border-border bg-background font-medium text-sm text-foreground transition-colors hover:bg-secondary cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {passkeyLoading ? (
                <Loader2 className="size-4 animate-spin" />
              ) : (
                <Fingerprint className="size-4" aria-hidden="true" />
              )}
              {t("passkey.button")}
            </button>
          </div>
        )}
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

