"use client";

import { useEffect, useState } from "react";
import { CheckCircle2, Loader2, Store, XCircle } from "lucide-react";
import { sellerService } from "@/modules/seller/api/sellerService";
import { ApiError } from "@/shared/api/client";
import { useSellerContext } from "./sellerContext";
import { useRouter } from "@/shared/i18n/navigation";
import { apiErrorMessage, notify } from "@/shared/lib/notify";
import { normalizeSlug, validateSlug } from "@/modules/seller/lib/slug";
import { useTranslations } from "next-intl";

type SlugStatus = "idle" | "checking" | "available" | "unavailable";

export function SellerRegister() {
  const t = useTranslations("seller");
  const { seller, refreshSeller } = useSellerContext();
  const router = useRouter();

  const [storeName, setStoreName] = useState(seller?.storeName ?? "");
  const [slug, setSlug] = useState(seller?.slug ?? "");
  const [description, setDescription] = useState(seller?.description ?? "");
  const [submitting, setSubmitting] = useState(false);
  const [slugStatus, setSlugStatus] = useState<SlugStatus>("idle");
  const [slugReason, setSlugReason] = useState<string | null>(null);

  const reapplication = seller?.status === "rejected";

  useEffect(() => {
    const normalized = normalizeSlug(slug);
    if (!normalized) {
      setSlugStatus("idle");
      setSlugReason(null);
      return;
    }

    const validation = validateSlug(normalized);
    if (validation.reason) {
      setSlugStatus("unavailable");
      setSlugReason(validation.reason);
      return;
    }

    let cancelled = false;
    setSlugStatus("checking");
    setSlugReason(null);
    const timer = window.setTimeout(() => {
      void sellerService
        .checkSlugAvailability(normalized)
        .then((response) => {
          if (cancelled) return;
          setSlugStatus(response.data.available ? "available" : "unavailable");
          setSlugReason(response.data.reason ?? null);
          if (response.data.available) setSlug(normalized);
        })
        .catch((error: unknown) => {
          if (cancelled) return;
          setSlugStatus("unavailable");
          setSlugReason(apiErrorMessage(error, t("register.slugCheckFailed")));
        });
    }, 400);

    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [slug, t]);

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    const validation = validateSlug(slug);
    if (validation.reason) {
      setSlugStatus("unavailable");
      setSlugReason(validation.reason);
      return;
    }
    if (slugStatus === "checking") return;

    setSubmitting(true);
    try {
      await sellerService.register({
        storeName: storeName.trim(),
        slug: validation.slug,
        description: description.trim() || undefined,
      });
      notify.success(
        t(reapplication ? "register.reapplied" : "register.success"),
      );
      await refreshSeller();
      router.replace("/seller");
    } catch (error) {
      if (error instanceof ApiError && error.code === "SELLER_SLUG_TAKEN") {
        setSlugStatus("unavailable");
        setSlugReason(t("register.slugTaken"));
      }
      notify.error(apiErrorMessage(error, t("register.submitFailed")));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="mx-auto w-full max-w-xl px-4 py-8">
      <div className="space-y-6 rounded-2xl border border-border bg-background p-6 shadow-xs">
        <div>
          <div className="flex items-center gap-2">
            <div className="flex size-10 items-center justify-center rounded-2xl bg-accent-brand/10 text-accent-brand">
              <Store className="size-5" />
            </div>
            <h1 className="text-2xl font-bold text-foreground">
              {t(reapplication ? "register.reapplyTitle" : "register.title")}
            </h1>
          </div>
          <p className="mt-2 text-sm text-muted-foreground">
            {t(reapplication ? "register.reapplySubtitle" : "register.subtitle")}
          </p>
        </div>

        {reapplication ? (
          <div className="rounded-xl border border-border bg-muted/30 p-3 text-xs text-muted-foreground">
            <div className="font-medium text-foreground">
              {t("register.reapplyReason")}
            </div>
            <p className="mt-1">{seller?.statusReason || t("statusPages.reasonFallback")}</p>
          </div>
        ) : (
          <p className="rounded-xl border border-border bg-secondary/30 p-3 text-xs text-muted-foreground">
            {t("register.pendingNote")}
          </p>
        )}

        <form onSubmit={(event) => void handleSubmit(event)} className="space-y-4">
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-foreground" htmlFor="seller-store-name">
              {t("register.storeName")}
            </label>
            <input
              id="seller-store-name"
              type="text"
              required
              minLength={2}
              maxLength={150}
              value={storeName}
              onChange={(event) => setStoreName(event.target.value)}
              placeholder={t("register.storeNamePlaceholder")}
              className="w-full rounded-xl border border-border bg-secondary/30 px-3 py-2 text-base font-bold text-foreground"
            />
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-foreground" htmlFor="seller-slug">
              {t("register.slug")}
            </label>
            <div className="relative">
              <span className="pointer-events-none absolute inset-y-0 start-3 flex items-center text-sm text-muted-foreground">
                /store/
              </span>
              <input
                id="seller-slug"
                type="text"
                required
                minLength={3}
                maxLength={63}
                dir="ltr"
                value={slug}
                onChange={(event) => setSlug(event.target.value)}
                onBlur={() => setSlug(normalizeSlug(slug))}
                placeholder={t("register.slugPlaceholder")}
                aria-invalid={slugStatus === "unavailable"}
                aria-describedby="seller-slug-status"
                className="w-full rounded-xl border border-border bg-secondary/30 py-2 pe-3 ps-16 text-base font-bold text-foreground"
              />
            </div>
            <div
              id="seller-slug-status"
              className={`flex items-start gap-1.5 text-[11px] ${
                slugStatus === "unavailable" ? "text-destructive" : "text-muted-foreground"
              }`}
            >
              {slugStatus === "checking" ? <Loader2 className="mt-0.5 size-3 animate-spin" /> : null}
              {slugStatus === "available" ? (
                <CheckCircle2 className="mt-0.5 size-3 text-emerald-600" />
              ) : null}
              {slugStatus === "unavailable" ? <XCircle className="mt-0.5 size-3" /> : null}
              <span>
                {slugStatus === "checking"
                  ? t("register.slugChecking")
                  : slugStatus === "available"
                    ? t("register.slugAvailable")
                    : slugReason || t("register.slugHint")}
              </span>
            </div>
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-foreground" htmlFor="seller-description">
              {t("register.description")}
            </label>
            <textarea
              id="seller-description"
              value={description}
              onChange={(event) => setDescription(event.target.value)}
              placeholder={t("register.descriptionPlaceholder")}
              rows={4}
              maxLength={2000}
              className="w-full rounded-xl border border-border bg-secondary/30 px-3 py-2 text-base font-bold text-foreground"
            />
          </div>

          <button
            type="submit"
            disabled={submitting || slugStatus === "checking"}
            className="inline-flex items-center justify-center gap-2 rounded-xl bg-accent-brand px-5 py-2.5 text-xs font-semibold text-white shadow-xs transition-all hover:bg-accent-brand/90 disabled:opacity-60"
          >
            {submitting && <Loader2 className="size-4 animate-spin" />}
            <span>{t(reapplication ? "register.reapplySubmit" : "register.submit")}</span>
          </button>
        </form>
      </div>
    </div>
  );
}
