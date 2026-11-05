"use client";

import { useState } from "react";
import { Loader2, Store } from "lucide-react";
import { vendorService } from "@/modules/vendor/api/vendorService";
import { useVendorContext } from "./vendorContext";
import { apiErrorMessage, notify } from "@/shared/lib/notify";
import { useTranslations } from "next-intl";

export function VendorRegister() {
  const t = useTranslations("vendor");
  const { refreshVendor } = useVendorContext();

  const [storeName, setStoreName] = useState("");
  const [slug, setSlug] = useState("");
  const [description, setDescription] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setSubmitting(true);
    try {
      await vendorService.register({
        storeName: storeName.trim(),
        slug: slug.trim(),
        description: description.trim() || undefined,
      });
      notify.success(t("register.success"));
      await refreshVendor();
    } catch (error) {
      notify.error(apiErrorMessage(error, t("register.submitFailed")));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="max-w-xl mx-auto w-full">
      <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-6">
        <div>
          <div className="flex items-center gap-2">
            <div className="size-10 rounded-2xl bg-accent-brand/10 text-accent-brand flex items-center justify-center">
              <Store className="size-5" />
            </div>
            <h1 className="text-2xl font-bold text-foreground">{t("register.title")}</h1>
          </div>
          <p className="text-sm text-muted-foreground mt-2">{t("register.subtitle")}</p>
        </div>

        <p className="text-xs text-muted-foreground bg-secondary/30 rounded-xl border border-border p-3">
          {t("register.pendingNote")}
        </p>

        <form onSubmit={(event) => void handleSubmit(event)} className="space-y-4">
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-foreground" htmlFor="vendor-store-name">
              {t("register.storeName")}
            </label>
            <input
              id="vendor-store-name"
              type="text"
              required
              minLength={2}
              maxLength={150}
              value={storeName}
              onChange={(event) => setStoreName(event.target.value)}
              placeholder={t("register.storeNamePlaceholder")}
              className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-base font-bold text-foreground"
            />
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-foreground" htmlFor="vendor-slug">
              {t("register.slug")}
            </label>
            <input
              id="vendor-slug"
              type="text"
              required
              maxLength={150}
              value={slug}
              onChange={(event) => setSlug(event.target.value)}
              placeholder={t("register.slugPlaceholder")}
              className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-base font-bold text-foreground"
            />
            <p className="text-[11px] text-muted-foreground">{t("register.slugHint")}</p>
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-foreground" htmlFor="vendor-description">
              {t("register.description")}
            </label>
            <textarea
              id="vendor-description"
              value={description}
              onChange={(event) => setDescription(event.target.value)}
              placeholder={t("register.descriptionPlaceholder")}
              rows={4}
              maxLength={2000}
              className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-base font-bold text-foreground"
            />
          </div>

          <button
            type="submit"
            disabled={submitting}
            className="inline-flex items-center justify-center gap-2 px-5 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all disabled:opacity-60"
          >
            {submitting && <Loader2 className="size-4 animate-spin" />}
            <span>{t("register.submit")}</span>
          </button>
        </form>
      </div>
    </div>
  );
}
