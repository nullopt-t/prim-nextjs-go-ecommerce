"use client";

import { useState, useEffect } from "react";
import { useAuthContext } from "@/modules/auth/model/authContext";
import { userService } from "@/modules/account/api/userService";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { useTranslations } from "next-intl";

export default function PersonalInfoForm() {
  const { user, refreshUser } = useAuthContext();
  const t = useTranslations("settings.personalInfo");

  const [name, setName] = useState(user?.name || "");
  const [phone, setPhone] = useState(user?.phone || "");
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    if (user) {
      setName(user.name || "");
      setPhone(user.phone || "");
    }
  }, [user]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSaving(true);
    try {
      await userService.updateProfile({
        fullName: name.trim(),
        phone: phone.trim() || undefined,
      });
      await refreshUser();
      notify.success(t("saveChanges"));
    } catch (err) {
      notify.error(apiErrorMessage(err, "Failed to update profile"));
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
