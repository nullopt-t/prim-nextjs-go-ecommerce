"use client";

import { useState, useEffect } from "react";
import { AlertCircle } from "lucide-react";
import { useAuthContext } from "@/modules/auth/model/authContext";
import { userService } from "@/modules/account/api/userService";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { useTranslations } from "next-intl";
import { isValidEgyptianPhone, normalizeEgyptianPhone } from "@/modules/account/model/addressSchema";

export default function PersonalInfoForm() {
  const { user, refreshUser } = useAuthContext();
  const t = useTranslations("settings.personalInfo");

  const [name, setName] = useState(user?.name || "");
  const [phone, setPhone] = useState(user?.phone || "");
  const [formError, setFormError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    if (user) {
      setName(user.name || "");
      setPhone(user.phone || "");
    }
  }, [user]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);
    if (!name.trim()) {
      const msg = t("invalidName");
      setFormError(msg);
      notify.error(msg);
      return;
    }
    if (phone.trim() && !isValidEgyptianPhone(phone)) {
      const msg = t("invalidPhone");
      setFormError(msg);
      notify.error(msg);
      return;
    }
    setIsSaving(true);
    try {
      await userService.updateProfile({
        fullName: name.trim(),
        phone: phone.trim() ? normalizeEgyptianPhone(phone) : undefined,
      });
      await refreshUser();
      setFormError(null);
    } catch (err) {
      const msg = apiErrorMessage(err, "Failed to update profile");
      setFormError(msg);
      notify.error(msg);
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="flex flex-col gap-8 w-full">
      <form className="flex flex-col gap-5" onSubmit={handleSubmit}>
        {/* Display Name */}
        <div className="flex flex-col gap-1.5">
          <label htmlFor="display-name" className="text-xs font-medium text-foreground">
            {t("displayName")}
          </label>
          <input
            id="display-name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={t("displayName")}
            className="h-10 w-full rounded-xl border border-border bg-background px-3.5 text-xs sm:text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-accent-brand transition"
          />
        </div>

        {/* Email */}
        <div className="flex flex-col gap-1.5">
          <label htmlFor="email" className="text-xs font-medium text-foreground">
            {t("email")}
          </label>
          <input
            id="email"
            type="email"
            value={user?.email || ""}
            readOnly
            disabled
            className="h-10 w-full rounded-xl border border-border bg-secondary/30 px-3.5 text-xs sm:text-sm text-foreground placeholder:text-muted-foreground disabled:opacity-75 disabled:cursor-not-allowed focus:outline-none transition"
          />
          <p className="text-[11px] text-muted-foreground">
            {t("emailNotice")}
          </p>
        </div>

        {/* Phone */}
        <div className="flex flex-col gap-1.5">
          <label htmlFor="phone" className="text-xs font-medium text-foreground">
            {t("phone")}
          </label>
          <input
            id="phone"
            type="tel"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            placeholder={t("notSet")}
            className="h-10 w-full rounded-xl border border-border bg-background px-3.5 text-xs sm:text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-accent-brand transition"
          />
        </div>

        {/* Submit button */}
        {formError && (
          <div
            role="alert"
            className="rounded-xl border border-destructive/30 bg-destructive/10 p-3 text-xs font-semibold text-destructive flex items-center gap-2"
          >
            <AlertCircle className="size-4 shrink-0" aria-hidden="true" />
            <span>{formError}</span>
          </div>
        )}

        <div className="pt-2">
          <button
            type="submit"
            disabled={isSaving}
            className="px-5 py-2.5 rounded-xl bg-accent-brand text-white text-xs sm:text-sm font-semibold hover:opacity-90 transition disabled:opacity-50 cursor-pointer"
          >
            {isSaving ? "..." : t("saveChanges")}
          </button>
        </div>
      </form>
    </div>
  );
}
