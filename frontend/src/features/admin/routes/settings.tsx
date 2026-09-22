import { useState } from "react";
import { Save, Key, Truck, Store } from "lucide-react";
import { INITIAL_SETTINGS, getLocalAdminData, setLocalAdminData } from "../data/mockAdminData";
import { toast } from "sonner";

export function AdminSettings() {
  const [settings, setSettings] = useState(() => getLocalAdminData("settings", INITIAL_SETTINGS));

  const handleSave = (e) => {
    e.preventDefault();
    setLocalAdminData("settings", settings);
    toast.success("Store configuration saved successfully!");
  };

  return (
    <div className="space-y-6 max-w-4xl">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Store Configuration</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Global store metadata, currency preferences, tax calculations, and API security keys.
          </p>
        </div>
        <button
          onClick={handleSave}
          className="flex items-center gap-2 px-5 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all shrink-0"
        >
          <Save className="size-4" />
          <span>Save Changes</span>
        </button>
      </div>

      <form onSubmit={handleSave} className="space-y-6">
        {/* General Store Details */}
        <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
          <h3 className="font-bold text-base text-foreground flex items-center gap-2 border-b border-border pb-3">
            <Store className="size-5 text-accent-brand" />
            <span>Store Profile</span>
          </h3>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">Store Name</label>
              <input
                type="text"
                value={settings.storeName}
                onChange={(e) => setSettings({ ...settings, storeName: e.target.value })}
                className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
              />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">Customer Support Email</label>
              <input
                type="email"
                value={settings.storeEmail}
                onChange={(e) => setSettings({ ...settings, storeEmail: e.target.value })}
                className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
              />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">Base Currency</label>
              <input
                type="text"
                value={settings.currency}
                onChange={(e) => setSettings({ ...settings, currency: e.target.value })}
                className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
              />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">Timezone</label>
              <input
                type="text"
                value={settings.timezone}
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
            <span>Shipping & Tax Rates</span>
          </h3>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">Standard Tax Rate (%)</label>
              <input
                type="number"
                step="0.1"
                value={settings.taxRatePercent}
                onChange={(e) => setSettings({ ...settings, taxRatePercent: parseFloat(e.target.value) || 0 })}
                className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
              />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">Flat Shipping Rate ($)</label>
              <input
                type="number"
                step="0.01"
                value={settings.flatRateShipping}
                onChange={(e) => setSettings({ ...settings, flatRateShipping: parseFloat(e.target.value) || 0 })}
                className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
              />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">Free Shipping Minimum ($)</label>
              <input
                type="number"
                value={settings.freeShippingThreshold}
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
            <span>API Keys & Webhooks</span>
          </h3>

          <div className="space-y-3">
            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">Live Public API Key</label>
              <input
                type="password"
                value={settings.apiKey}
                onChange={(e) => setSettings({ ...settings, apiKey: e.target.value })}
                className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs font-mono text-foreground"
              />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">Stripe Webhook Signing Secret</label>
              <input
                type="password"
                value={settings.stripeWebhookSecret}
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
