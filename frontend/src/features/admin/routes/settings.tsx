"use client";

import { useState, useEffect } from "react";
import {
  Save,
  Store,
  Truck,
  CreditCard,
  Bell,
  ShieldCheck,
  Building2,
  Globe,
  Megaphone,
  Percent,
  Key,
  Mail,
  ShieldAlert,
  Eye,
  EyeOff
} from "lucide-react";
import { notify, apiErrorMessage } from "@/lib/notify";
import { adminService } from "@/services/admin";
import { useTranslations } from "next-intl";

type SettingsTab = "general" | "shipping" | "payments" | "notifications" | "security";

// Display fallbacks only. The backend returns the authoritative resolved settings
// (persisted overrides merged onto server defaults) which always win on load.
const DEFAULT_SETTINGS: Record<string, any> = {
  // Store Identity & Localization
  storeName: "PRIM Superstore",
  storeEmail: "support@prim.com",
  storePhone: "+20 100 000 0000",
  storeAddress: "Smart Village, Building B12, Giza, Egypt",
  currency: "EGP",
  timezone: "Africa/Cairo",
  defaultLocale: "ar",
  announcementEnabled: true,
  announcementText: "Welcome to PRIM! Enjoy free delivery on qualifying orders.",

  // Shipping & Delivery
  flatRateShipping: 9.99,
  freeShippingThreshold: 150,
  estimatedDeliveryDays: "2 - 4 business days",

  // Payments & Taxes
  allowCod: true,
  codFee: 0,
  allowCards: true,
  taxRatePercent: 0,
  pricesIncludeTax: true,
  cartReservationMinutes: 15,

  // Notifications
  notifyOrderConfirmation: true,
  notifyOrderShipped: true,
  adminNotificationEmail: "operations@prim.com",
  notifyLowStock: true,

  // Security & Integrations
  apiKey: "",
  stripeWebhookSecret: "",
  maintenanceMode: false,
  maintenanceMessage: "PRIM is currently undergoing scheduled maintenance to upgrade our warehouse systems. We will be back shortly!"
};

export function AdminSettings() {
  const t = useTranslations("admin");
  const [activeTab, setActiveTab] = useState<SettingsTab>("general");
  const [settings, setSettings] = useState<Record<string, any>>({ ...DEFAULT_SETTINGS });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [showApiKey, setShowApiKey] = useState(false);
  const [showWebhookSecret, setShowWebhookSecret] = useState(false);

  useEffect(() => {
    let isMounted = true;
    const loadSettings = async () => {
      try {
        const res = await adminService.getSettings();
        const data = res?.data || {};
        if (isMounted) {
          setSettings({ ...DEFAULT_SETTINGS, ...data });
        }
      } catch (err) {
        if (isMounted) {
          notify.error(apiErrorMessage(err, t("settings.toast.loadFailed")));
        }
      } finally {
        if (isMounted) setLoading(false);
      }
    };
    loadSettings();
    return () => {
      isMounted = false;
    };
  }, [t]);

  const handleSave = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    setSaving(true);
    try {
      await adminService.updateSettings(settings);
      notify.success(t("settings.toast.saved"));
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("settings.toast.saveFailed")));
    } finally {
      setSaving(false);
    }
  };


  const tabs = [
    { id: "general" as const, label: t("settings.tabs.general"), icon: Store },
    { id: "shipping" as const, label: t("settings.tabs.shipping"), icon: Truck },
    { id: "payments" as const, label: t("settings.tabs.payments"), icon: CreditCard },
    { id: "notifications" as const, label: t("settings.tabs.notifications"), icon: Bell },
    { id: "security" as const, label: t("settings.tabs.security"), icon: ShieldCheck }
  ];

  return (
    <div className="space-y-6 w-full">
      {/* Top Header Card */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-3xl border border-border/80 shadow-xs">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="text-2xl font-bold tracking-tight text-foreground">{t("settings.title")}</h1>
            <span className="px-2.5 py-0.5 rounded-full text-[11px] font-bold bg-accent-brand/10 text-accent-brand border border-accent-brand/20">
              v2.4
            </span>
          </div>
          <p className="text-xs text-muted-foreground mt-1">{t("settings.subtitle")}</p>
        </div>

        <button
          onClick={() => handleSave()}
          disabled={saving || loading}
          className="flex items-center gap-2 px-5 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all shrink-0 disabled:opacity-50"
        >
          <Save className="size-4" />
          <span>{saving ? t("settings.saving") : t("settings.saveChanges")}</span>
        </button>
      </div>

      {/* Navigation Tab Pills */}
      <div className="flex items-center gap-2 overflow-x-auto pb-1 scrollbar-none">
        {tabs.map((tab) => {
          const Icon = tab.icon;
          const isActive = activeTab === tab.id;
          return (
            <button
              key={tab.id}
              type="button"
              onClick={() => setActiveTab(tab.id)}
              className={`flex items-center gap-2 px-4 py-2.5 rounded-2xl text-xs font-semibold whitespace-nowrap transition-all border ${
                isActive
                  ? "bg-background text-foreground border-border/80 shadow-xs font-bold"
                  : "bg-secondary/20 text-muted-foreground border-transparent hover:text-foreground hover:bg-secondary/40"
              }`}
            >
              <Icon className={`size-4 ${isActive ? "text-accent-brand" : "text-muted-foreground"}`} />
              <span>{tab.label}</span>
            </button>
          );
        })}
      </div>

      {/* Main Tab Content */}
      <div className="space-y-6">
        {/* ─── TAB 1: GENERAL PROFILE ─── */}
        {activeTab === "general" && (
          <div className="space-y-6">
            {/* Store Identity */}
            <div className="bg-background border border-border/80 rounded-2xl p-6 shadow-xs space-y-5">
              <div className="border-b border-border/60 pb-3">
                <h3 className="font-bold text-sm text-foreground flex items-center gap-2">
                  <Building2 className="size-4 text-accent-brand" />
                  <span>{t("settings.sections.storeIdentity")}</span>
                </h3>
                <p className="text-[11px] text-muted-foreground mt-0.5">
                  {t("settings.sections.storeIdentityDesc")}
                </p>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-foreground">
                    {t("settings.fields.storeName")}
                  </label>
                  <input
                    type="text"
                    value={settings.storeName || ""}
                    onChange={(e) => setSettings({ ...settings, storeName: e.target.value })}
                    className="w-full px-3.5 py-2.5 bg-secondary/20 border border-border rounded-xl text-xs text-foreground focus:ring-2 focus:ring-accent-brand/20 focus:border-accent-brand focus:outline-hidden"
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-foreground">
                    {t("settings.fields.storeEmail")}
                  </label>
                  <input
                    type="email"
                    value={settings.storeEmail || ""}
                    onChange={(e) => setSettings({ ...settings, storeEmail: e.target.value })}
                    className="w-full px-3.5 py-2.5 bg-secondary/20 border border-border rounded-xl text-xs text-foreground focus:ring-2 focus:ring-accent-brand/20 focus:border-accent-brand focus:outline-hidden"
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-foreground">
                    {t("settings.fields.storePhone")}
                  </label>
                  <input
                    type="text"
                    value={settings.storePhone || ""}
                    onChange={(e) => setSettings({ ...settings, storePhone: e.target.value })}
                    className="w-full px-3.5 py-2.5 bg-secondary/20 border border-border rounded-xl text-xs text-foreground focus:ring-2 focus:ring-accent-brand/20 focus:border-accent-brand focus:outline-hidden"
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-foreground">
                    {t("settings.fields.storeAddress")}
                  </label>
                  <input
                    type="text"
                    value={settings.storeAddress || ""}
                    onChange={(e) => setSettings({ ...settings, storeAddress: e.target.value })}
                    className="w-full px-3.5 py-2.5 bg-secondary/20 border border-border rounded-xl text-xs text-foreground focus:ring-2 focus:ring-accent-brand/20 focus:border-accent-brand focus:outline-hidden"
                  />
                </div>
              </div>
            </div>

            {/* Regional & Localization */}
            <div className="bg-background border border-border/80 rounded-2xl p-6 shadow-xs space-y-5">
              <div className="border-b border-border/60 pb-3">
                <h3 className="font-bold text-sm text-foreground flex items-center gap-2">
                  <Globe className="size-4 text-accent-brand" />
                  <span>{t("settings.sections.regional")}</span>
                </h3>
                <p className="text-[11px] text-muted-foreground mt-0.5">
                  {t("settings.sections.regionalDesc")}
                </p>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-foreground">
                    {t("settings.fields.currency")}
                  </label>
                  <select
                    value={settings.currency || "EGP"}
                    onChange={(e) => setSettings({ ...settings, currency: e.target.value })}
                    className="w-full px-3.5 py-2.5 bg-secondary/20 border border-border rounded-xl text-xs text-foreground focus:ring-2 focus:ring-accent-brand/20 focus:border-accent-brand focus:outline-hidden"
                  >
                    <option value="EGP">EGP - Egyptian Pound (ج.م)</option>
                    <option value="USD">USD - US Dollar ($)</option>
                    <option value="SAR">SAR - Saudi Riyal (ر.س)</option>
                    <option value="AED">AED - UAE Dirham (د.إ)</option>
                  </select>
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-foreground">
                    {t("settings.fields.timezone")}
                  </label>
                  <select
                    value={settings.timezone || "Africa/Cairo"}
                    onChange={(e) => setSettings({ ...settings, timezone: e.target.value })}
                    className="w-full px-3.5 py-2.5 bg-secondary/20 border border-border rounded-xl text-xs text-foreground focus:ring-2 focus:ring-accent-brand/20 focus:border-accent-brand focus:outline-hidden"
                  >
                    <option value="Africa/Cairo">Africa/Cairo (UTC+2)</option>
                    <option value="Asia/Riyadh">Asia/Riyadh (UTC+3)</option>
                    <option value="Asia/Dubai">Asia/Dubai (UTC+4)</option>
                    <option value="UTC">Coordinated Universal Time (UTC)</option>
                  </select>
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-foreground">
                    {t("settings.fields.defaultLocale")}
                  </label>
                  <select
                    value={settings.defaultLocale || "ar"}
                    onChange={(e) => setSettings({ ...settings, defaultLocale: e.target.value })}
                    className="w-full px-3.5 py-2.5 bg-secondary/20 border border-border rounded-xl text-xs text-foreground focus:ring-2 focus:ring-accent-brand/20 focus:border-accent-brand focus:outline-hidden"
                  >
                    <option value="ar">العربية (Arabic - RTL)</option>
                    <option value="en">English (LTR)</option>
                  </select>
                </div>
              </div>
            </div>

            {/* Announcement Ribbon */}
            <div className="bg-background border border-border/80 rounded-2xl p-6 shadow-xs space-y-4">
              <div className="border-b border-border/60 pb-3 flex items-center justify-between">
                <div>
                  <h3 className="font-bold text-sm text-foreground flex items-center gap-2">
                    <Megaphone className="size-4 text-accent-brand" />
                    <span>{t("settings.sections.announcement")}</span>
                  </h3>
                  <p className="text-[11px] text-muted-foreground mt-0.5">
                    {t("settings.sections.announcementDesc")}
                  </p>
                </div>

                <label className="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    checked={Boolean(settings.announcementEnabled)}
                    onChange={(e) => setSettings({ ...settings, announcementEnabled: e.target.checked })}
                    className="sr-only peer"
                  />
                  <div className="w-10 h-6 bg-secondary/60 peer-focus:outline-hidden rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-border after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-accent-brand"></div>
                </label>
              </div>

              {settings.announcementEnabled && (
                <div className="space-y-1.5 pt-1">
                  <label className="text-xs font-semibold text-foreground">
                    {t("settings.fields.announcementText")}
                  </label>
                  <input
                    type="text"
                    value={settings.announcementText || ""}
                    onChange={(e) => setSettings({ ...settings, announcementText: e.target.value })}
                    className="w-full px-3.5 py-2.5 bg-secondary/20 border border-border rounded-xl text-xs text-foreground focus:ring-2 focus:ring-accent-brand/20 focus:border-accent-brand focus:outline-hidden"
                  />
                </div>
              )}
            </div>
          </div>
        )}

        {/* ─── TAB 2: SHIPPING & DELIVERY ─── */}
        {activeTab === "shipping" && (
          <div className="space-y-6">
            <div className="bg-background border border-border/80 rounded-2xl p-6 shadow-xs space-y-5">
              <div className="border-b border-border/60 pb-3">
                <h3 className="font-bold text-sm text-foreground flex items-center gap-2">
                  <Truck className="size-4 text-accent-brand" />
                  <span>{t("settings.sections.deliveryRates")}</span>
                </h3>
                <p className="text-[11px] text-muted-foreground mt-0.5">
                  {t("settings.sections.deliveryRatesDesc")}
                </p>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-foreground">
                    {t("settings.fields.flatShipping")}
                  </label>
                  <input
                    type="number"
                    step="0.5"
                    min="0"
                    value={settings.flatRateShipping ?? ""}
                    onChange={(e) =>
                      setSettings({ ...settings, flatRateShipping: parseFloat(e.target.value) || 0 })
                    }
                    className="w-full px-3.5 py-2.5 bg-secondary/20 border border-border rounded-xl text-xs text-foreground focus:ring-2 focus:ring-accent-brand/20 focus:border-accent-brand focus:outline-hidden"
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-foreground">
                    {t("settings.fields.freeShippingMin")}
                  </label>
                  <input
                    type="number"
                    step="10"
                    min="0"
                    value={settings.freeShippingThreshold ?? ""}
                    onChange={(e) =>
                      setSettings({ ...settings, freeShippingThreshold: parseFloat(e.target.value) || 0 })
                    }
                    className="w-full px-3.5 py-2.5 bg-secondary/20 border border-border rounded-xl text-xs text-foreground focus:ring-2 focus:ring-accent-brand/20 focus:border-accent-brand focus:outline-hidden"
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-foreground">
                    {t("settings.fields.estimatedDeliveryDays")}
                  </label>
                  <input
                    type="text"
                    value={settings.estimatedDeliveryDays || ""}
                    onChange={(e) =>
                      setSettings({ ...settings, estimatedDeliveryDays: e.target.value })
                    }
                    className="w-full px-3.5 py-2.5 bg-secondary/20 border border-border rounded-xl text-xs text-foreground focus:ring-2 focus:ring-accent-brand/20 focus:border-accent-brand focus:outline-hidden"
                  />
                </div>
              </div>
            </div>
          </div>
        )}

        {/* ─── TAB 3: PAYMENTS & TAX ─── */}
        {activeTab === "payments" && (
          <div className="space-y-6">
            {/* Payment Channels */}
            <div className="bg-background border border-border/80 rounded-2xl p-6 shadow-xs space-y-5">
              <div className="border-b border-border/60 pb-3">
                <h3 className="font-bold text-sm text-foreground flex items-center gap-2">
                  <CreditCard className="size-4 text-accent-brand" />
                  <span>{t("settings.sections.paymentMethods")}</span>
                </h3>
                <p className="text-[11px] text-muted-foreground mt-0.5">
                  {t("settings.sections.paymentMethodsDesc")}
                </p>
              </div>

              <div className="space-y-4">
                {/* Cash on Delivery */}
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4 rounded-xl border border-border/70 bg-secondary/10">
                  <div>
                    <h4 className="text-xs font-bold text-foreground">
                      {t("settings.fields.allowCod")}
                    </h4>
                    <p className="text-[11px] text-muted-foreground">
                      Accept physical cash collection upon package delivery.
                    </p>
                  </div>
                  <label className="relative inline-flex items-center cursor-pointer shrink-0">
                    <input
                      type="checkbox"
                      checked={Boolean(settings.allowCod)}
                      onChange={(e) => setSettings({ ...settings, allowCod: e.target.checked })}
                      className="sr-only peer"
                    />
                    <div className="w-10 h-6 bg-secondary/60 peer-focus:outline-hidden rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-border after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-accent-brand"></div>
                  </label>
                </div>

                {settings.allowCod && (
                  <div className="ps-4 pe-2 py-2 border-s-2 border-accent-brand max-w-sm space-y-1.5">
                    <label className="text-xs font-semibold text-foreground">
                      {t("settings.fields.codFee")}
                    </label>
                    <input
                      type="number"
                      step="1"
                      min="0"
                      value={settings.codFee ?? ""}
                      onChange={(e) => setSettings({ ...settings, codFee: parseFloat(e.target.value) || 0 })}
                      className="w-full px-3.5 py-2 bg-secondary/20 border border-border rounded-xl text-xs text-foreground focus:ring-2 focus:ring-accent-brand/20 focus:border-accent-brand focus:outline-hidden"
                    />
                  </div>
                )}

                {/* Cards */}
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4 rounded-xl border border-border/70 bg-secondary/10">
                  <div>
                    <h4 className="text-xs font-bold text-foreground">
                      {t("settings.fields.allowCards")}
                    </h4>
                    <p className="text-[11px] text-muted-foreground">
                      Accept Visa, Mastercard, Meeza, and mobile wallets.
                    </p>
                  </div>
                  <label className="relative inline-flex items-center cursor-pointer shrink-0">
                    <input
                      type="checkbox"
                      checked={Boolean(settings.allowCards)}
                      onChange={(e) => setSettings({ ...settings, allowCards: e.target.checked })}
                      className="sr-only peer"
                    />
                    <div className="w-10 h-6 bg-secondary/60 peer-focus:outline-hidden rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-border after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-accent-brand"></div>
                  </label>
                </div>
              </div>
            </div>

            {/* Taxes & Reservation */}
            <div className="bg-background border border-border/80 rounded-2xl p-6 shadow-xs space-y-5">
              <div className="border-b border-border/60 pb-3">
                <h3 className="font-bold text-sm text-foreground flex items-center gap-2">
                  <Percent className="size-4 text-accent-brand" />
                  <span>{t("settings.sections.taxRules")}</span>
                </h3>
                <p className="text-[11px] text-muted-foreground mt-0.5">
                  {t("settings.sections.taxRulesDesc")}
                </p>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-foreground">
                    {t("settings.fields.taxRate")}
                  </label>
                  <input
                    type="number"
                    step="0.5"
                    min="0"
                    max="100"
                    value={settings.taxRatePercent ?? ""}
                    onChange={(e) =>
                      setSettings({ ...settings, taxRatePercent: parseFloat(e.target.value) || 0 })
                    }
                    className="w-full px-3.5 py-2.5 bg-secondary/20 border border-border rounded-xl text-xs text-foreground focus:ring-2 focus:ring-accent-brand/20 focus:border-accent-brand focus:outline-hidden"
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-foreground">
                    {t("settings.fields.cartReservationMinutes")}
                  </label>
                  <input
                    type="number"
                    step="1"
                    min="5"
                    max="120"
                    value={settings.cartReservationMinutes ?? ""}
                    onChange={(e) =>
                      setSettings({
                        ...settings,
                        cartReservationMinutes: parseInt(e.target.value, 10) || 15
                      })
                    }
                    className="w-full px-3.5 py-2.5 bg-secondary/20 border border-border rounded-xl text-xs text-foreground focus:ring-2 focus:ring-accent-brand/20 focus:border-accent-brand focus:outline-hidden"
                  />
                </div>
              </div>

              <div className="flex items-center gap-3 pt-2">
                <input
                  type="checkbox"
                  id="pricesIncludeTax"
                  checked={Boolean(settings.pricesIncludeTax)}
                  onChange={(e) => setSettings({ ...settings, pricesIncludeTax: e.target.checked })}
                  className="rounded-md border-border text-accent-brand focus:ring-accent-brand size-4 cursor-pointer"
                />
                <label
                  htmlFor="pricesIncludeTax"
                  className="text-xs font-medium text-foreground cursor-pointer select-none"
                >
                  {t("settings.fields.pricesIncludeTax")}
                </label>
              </div>
            </div>
          </div>
        )}

        {/* ─── TAB 4: NOTIFICATIONS & EMAIL ─── */}
        {activeTab === "notifications" && (
          <div className="space-y-6">
            {/* Customer Notifications */}
            <div className="bg-background border border-border/80 rounded-2xl p-6 shadow-xs space-y-5">
              <div className="border-b border-border/60 pb-3">
                <h3 className="font-bold text-sm text-foreground flex items-center gap-2">
                  <Mail className="size-4 text-accent-brand" />
                  <span>{t("settings.sections.emailNotifications")}</span>
                </h3>
                <p className="text-[11px] text-muted-foreground mt-0.5">
                  {t("settings.sections.emailNotificationsDesc")}
                </p>
              </div>

              <div className="space-y-3">
                <div className="flex items-center justify-between p-4 rounded-xl border border-border/70 bg-secondary/10">
                  <span className="text-xs font-semibold text-foreground">
                    {t("settings.fields.notifyOrderConfirmation")}
                  </span>
                  <label className="relative inline-flex items-center cursor-pointer shrink-0">
                    <input
                      type="checkbox"
                      checked={Boolean(settings.notifyOrderConfirmation)}
                      onChange={(e) =>
                        setSettings({ ...settings, notifyOrderConfirmation: e.target.checked })
                      }
                      className="sr-only peer"
                    />
                    <div className="w-10 h-6 bg-secondary/60 peer-focus:outline-hidden rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-border after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-accent-brand"></div>
                  </label>
                </div>

                <div className="flex items-center justify-between p-4 rounded-xl border border-border/70 bg-secondary/10">
                  <span className="text-xs font-semibold text-foreground">
                    {t("settings.fields.notifyOrderShipped")}
                  </span>
                  <label className="relative inline-flex items-center cursor-pointer shrink-0">
                    <input
                      type="checkbox"
                      checked={Boolean(settings.notifyOrderShipped)}
                      onChange={(e) =>
                        setSettings({ ...settings, notifyOrderShipped: e.target.checked })
                      }
                      className="sr-only peer"
                    />
                    <div className="w-10 h-6 bg-secondary/60 peer-focus:outline-hidden rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-border after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-accent-brand"></div>
                  </label>
                </div>
              </div>
            </div>

            {/* Admin Internal Alerts */}
            <div className="bg-background border border-border/80 rounded-2xl p-6 shadow-xs space-y-5">
              <div className="border-b border-border/60 pb-3">
                <h3 className="font-bold text-sm text-foreground flex items-center gap-2">
                  <Bell className="size-4 text-accent-brand" />
                  <span>{t("settings.sections.adminAlerts")}</span>
                </h3>
                <p className="text-[11px] text-muted-foreground mt-0.5">
                  {t("settings.sections.adminAlertsDesc")}
                </p>
              </div>

              <div className="space-y-4">
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-foreground">
                    {t("settings.fields.adminNotificationEmail")}
                  </label>
                  <input
                    type="email"
                    value={settings.adminNotificationEmail || ""}
                    onChange={(e) =>
                      setSettings({ ...settings, adminNotificationEmail: e.target.value })
                    }
                    className="w-full px-3.5 py-2.5 bg-secondary/20 border border-border rounded-xl text-xs text-foreground focus:ring-2 focus:ring-accent-brand/20 focus:border-accent-brand focus:outline-hidden"
                  />
                </div>

                <div className="flex items-center justify-between p-4 rounded-xl border border-border/70 bg-secondary/10">
                  <span className="text-xs font-semibold text-foreground">
                    {t("settings.fields.notifyLowStock")}
                  </span>
                  <label className="relative inline-flex items-center cursor-pointer shrink-0">
                    <input
                      type="checkbox"
                      checked={Boolean(settings.notifyLowStock)}
                      onChange={(e) => setSettings({ ...settings, notifyLowStock: e.target.checked })}
                      className="sr-only peer"
                    />
                    <div className="w-10 h-6 bg-secondary/60 peer-focus:outline-hidden rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-border after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-accent-brand"></div>
                  </label>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* ─── TAB 5: SECURITY & INTEGRATIONS ─── */}
        {activeTab === "security" && (
          <div className="space-y-6">
            {/* API Credentials */}
            <div className="bg-background border border-border/80 rounded-2xl p-6 shadow-xs space-y-5">
              <div className="border-b border-border/60 pb-3">
                <h3 className="font-bold text-sm text-foreground flex items-center gap-2">
                  <Key className="size-4 text-accent-brand" />
                  <span>{t("settings.sections.apiKeys")}</span>
                </h3>
                <p className="text-[11px] text-muted-foreground mt-0.5">
                  {t("settings.sections.apiKeysDesc")}
                </p>
              </div>

              <div className="space-y-4">
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-foreground flex items-center justify-between">
                    <span>{t("settings.fields.apiKey")}</span>
                    <button
                      type="button"
                      onClick={() => setShowApiKey(!showApiKey)}
                      className="text-muted-foreground hover:text-foreground text-[11px] flex items-center gap-1 font-medium"
                    >
                      {showApiKey ? <EyeOff className="size-3" /> : <Eye className="size-3" />}
                      <span>{showApiKey ? "Hide" : "Show"}</span>
                    </button>
                  </label>
                  <input
                    type={showApiKey ? "text" : "password"}
                    value={settings.apiKey || ""}
                    onChange={(e) => setSettings({ ...settings, apiKey: e.target.value })}
                    className="w-full px-3.5 py-2.5 bg-secondary/20 border border-border rounded-xl text-xs font-mono text-foreground focus:ring-2 focus:ring-accent-brand/20 focus:border-accent-brand focus:outline-hidden"
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-foreground flex items-center justify-between">
                    <span>{t("settings.fields.webhookSecret")}</span>
                    <button
                      type="button"
                      onClick={() => setShowWebhookSecret(!showWebhookSecret)}
                      className="text-muted-foreground hover:text-foreground text-[11px] flex items-center gap-1 font-medium"
                    >
                      {showWebhookSecret ? <EyeOff className="size-3" /> : <Eye className="size-3" />}
                      <span>{showWebhookSecret ? "Hide" : "Show"}</span>
                    </button>
                  </label>
                  <input
                    type={showWebhookSecret ? "text" : "password"}
                    value={settings.stripeWebhookSecret || ""}
                    onChange={(e) => setSettings({ ...settings, stripeWebhookSecret: e.target.value })}
                    className="w-full px-3.5 py-2.5 bg-secondary/20 border border-border rounded-xl text-xs font-mono text-foreground focus:ring-2 focus:ring-accent-brand/20 focus:border-accent-brand focus:outline-hidden"
                  />
                </div>
              </div>
            </div>

            {/* Maintenance Mode */}
            <div className="bg-background border border-amber-500/30 rounded-2xl p-6 shadow-xs space-y-4">
              <div className="border-b border-border/60 pb-3 flex items-center justify-between">
                <div>
                  <h3 className="font-bold text-sm text-foreground flex items-center gap-2">
                    <ShieldAlert className="size-4 text-amber-500" />
                    <span>{t("settings.sections.maintenance")}</span>
                  </h3>
                  <p className="text-[11px] text-muted-foreground mt-0.5">
                    {t("settings.sections.maintenanceDesc")}
                  </p>
                </div>

                <label className="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    checked={Boolean(settings.maintenanceMode)}
                    onChange={(e) => setSettings({ ...settings, maintenanceMode: e.target.checked })}
                    className="sr-only peer"
                  />
                  <div className="w-10 h-6 bg-secondary/60 peer-focus:outline-hidden rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-border after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-amber-500"></div>
                </label>
              </div>

              {settings.maintenanceMode && (
                <div className="space-y-1.5 pt-1">
                  <label className="text-xs font-semibold text-foreground">
                    {t("settings.fields.maintenanceMessage")}
                  </label>
                  <textarea
                    rows={3}
                    value={settings.maintenanceMessage || ""}
                    onChange={(e) => setSettings({ ...settings, maintenanceMessage: e.target.value })}
                    className="w-full px-3.5 py-2.5 bg-secondary/20 border border-border rounded-xl text-xs text-foreground focus:ring-2 focus:ring-accent-brand/20 focus:border-accent-brand focus:outline-hidden resize-none"
                  />
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}



