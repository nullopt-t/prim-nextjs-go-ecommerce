"use client";

import { useState, useEffect } from "react";
import { Save, Key, Truck, Store } from "lucide-react";
import { notify, apiErrorMessage } from "@/lib/notify";
import { adminService } from "@/services/admin";
import { useTranslations } from "next-intl";

const DEFAULT_SETTINGS: Record<string, any> = {
  storeName: "PRIM Superstore",
  storeEmail: "support@prim.com",
  currency: "EGP (E£)",
  timezone: "UTC",
  taxRatePercent: 8.5,
  flatRateShipping: 15.00,
  freeShippingThreshold: 150.00,
  apiKey: "pk_test_prim_live_publishable_key_9921",
  stripeWebhookSecret: "whsec_prim_live_webhook_secret_8812"
};

export function AdminSettings() {
  const t = useTranslations("admin");
  const [settings, setSettings] = useState<Record<string, any>>({ ...DEFAULT_SETTINGS });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    let isMounted = true;
    const loadSettings = async () => {
      try {
        const res = await adminService.getSettings();
        const data = res?.data || {};
        if (isMounted) {
          setSettings({ ...DEFAULT_SETTINGS, ...data });
        }
      } catch {
        // Fall back to defaults/localStorage if API is unavailable
        if (typeof window !== "undefined") {
          const saved = localStorage.getItem("prim_admin_settings");
          if (saved) {
            try {
              if (isMounted) setSettings({ ...DEFAULT_SETTINGS, ...JSON.parse(saved) });
            } catch {}
          }
        }
      } finally {
        if (isMounted) setLoading(false);
      }
    };
    loadSettings();
    return () => { isMounted = false; };
  }, []);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    try {
      await adminService.updateSettings(settings);
      if (typeof window !== "undefined") {
        localStorage.setItem("prim_admin_settings", JSON.stringify(settings));
      }
      notify.success(t("settings.toast.saved"));
    } catch (err: any) {
      // Persist locally even if API fails so changes aren't lost
      if (typeof window !== "undefined") {
        localStorage.setItem("prim_admin_settings", JSON.stringify(settings));
      }
      notify.error(apiErrorMessage(err, t("settings.toast.saveFailed")));
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="space-y-6 w-full">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t("settings.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {t("settings.subtitle")}
          </p>
        </div>
        <button
          onClick={handleSave}
          disabled={saving || loading}
          className="flex items-center gap-2 px-5 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all shrink-0 disabled:opacity-50"
        >
          <Save className="size-4" />
          <span>{saving ? t("settings.saving") : t("settings.saveChanges")}</span>
        </button>
      </div>

      <form onSubmit={handleSave} className="space-y-6">
        {/* General Store Details */}
        <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
          <h3 className="font-bold text-base text-foreground flex items-center gap-2 border-b border-border pb-3">
            <Store className="size-5 text-accent-brand" />
            <span>{t("settings.storeProfile")}</span>
          </h3>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">{t("settings.storeName")}</label>
              <input
                type="text"
                value={settings.storeName || ""}
                onChange={(e) => setSettings({ ...settings, storeName: e.target.value })}
                className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
              />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">{t("settings.supportEmail")}</label>
              <input
                type="email"
                value={settings.storeEmail || ""}
                onChange={(e) => setSettings({ ...settings, storeEmail: e.target.value })}
                className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
              />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">{t("settings.currency")}</label>
              <input
                type="text"
                value={settings.currency || ""}
                onChange={(e) => setSettings({ ...settings, currency: e.target.value })}
                className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
              />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">{t("settings.timezone")}</label>
              <input
                type="text"
                value={settings.timezone || ""}
                onChange={(e) => setSettings({ ...settings, timezone: e.target.value })}
                className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
              />
            </div>
          </div>
        </div>

        {/* Taxes & Shipping */}
        <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
          <h3 className="font-bold text-base text-foreground flex items-center gap-2 border-b border-border pb-3">
            <Truck className="size-5 text-accent-brand" />
            <span>{t("settings.shippingTax")}</span>
          </h3>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">{t("settings.taxRate")}</label>
              <input
                type="number"
                step="0.1"
                value={settings.taxRatePercent ?? ""}
                onChange={(e) => setSettings({ ...settings, taxRatePercent: parseFloat(e.target.value) || 0 })}
                className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
              />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">{t("settings.flatShipping")}</label>
              <input
                type="number"
                step="0.01"
                value={settings.flatRateShipping ?? ""}
                onChange={(e) => setSettings({ ...settings, flatRateShipping: parseFloat(e.target.value) || 0 })}
                className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
              />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">{t("settings.freeShippingMin")}</label>
              <input
                type="number"
                value={settings.freeShippingThreshold ?? ""}
                onChange={(e) => setSettings({ ...settings, freeShippingThreshold: parseFloat(e.target.value) || 0 })}
                className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
              />
            </div>
          </div>
        </div>

        {/* API Credentials */}
        <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
          <h3 className="font-bold text-base text-foreground flex items-center gap-2 border-b border-border pb-3">
            <Key className="size-5 text-accent-brand" />
            <span>{t("settings.apiKeys")}</span>
          </h3>

          <div className="space-y-3">
            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">{t("settings.apiKey")}</label>
              <input
                type="password"
                value={settings.apiKey || ""}
                onChange={(e) => setSettings({ ...settings, apiKey: e.target.value })}
                className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs font-mono text-foreground"
              />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">{t("settings.webhookSecret")}</label>
              <input
                type="password"
                value={settings.stripeWebhookSecret || ""}
                onChange={(e) => setSettings({ ...settings, stripeWebhookSecret: e.target.value })}
                className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs font-mono text-foreground"
              />
            </div>
          </div>
        </div>
      </form>
    </div>
  );
}
