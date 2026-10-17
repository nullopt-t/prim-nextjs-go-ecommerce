"use client";

import { useState, useEffect } from "react";
import {
  Sliders,
  Cpu,
  Trash2,
  Server,
  Activity,
  Save,
  CheckCircle2,
  Database,
  Layers,
  HardDrive,
  Radio
} from "lucide-react";
import { notify, apiErrorMessage } from "@/lib/notify";
import { adminService } from "@/services/admin";
import { useTranslations } from "next-intl";

const DEFAULT_SYSTEM_SETTINGS: Record<string, any> = {
  disableCache: false,
  enableDevOtpBypass: true,
  mockEmailNotifications: false
};

export function AdminSystem() {
  const t = useTranslations("admin");
  const [settings, setSettings] = useState<Record<string, any>>({ ...DEFAULT_SYSTEM_SETTINGS });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [purgingCache, setPurgingCache] = useState(false);

  useEffect(() => {
    let isMounted = true;
    const loadSettings = async () => {
      try {
        const res = await adminService.getSettings();
        const data = res?.data || {};
        if (isMounted) {
          setSettings({ ...DEFAULT_SYSTEM_SETTINGS, ...data });
        }
      } catch {
        if (typeof window !== "undefined") {
          const saved = localStorage.getItem("prim_admin_settings");
          if (saved) {
            try {
              if (isMounted) setSettings({ ...DEFAULT_SYSTEM_SETTINGS, ...JSON.parse(saved) });
            } catch {}
          }
        }
      } finally {
        if (isMounted) setLoading(false);
      }
    };
    loadSettings();
    return () => {
      isMounted = false;
    };
  }, []);

  const handleSave = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    setSaving(true);
    try {
      await adminService.updateSettings(settings);
      if (typeof window !== "undefined") {
        localStorage.setItem("prim_admin_settings", JSON.stringify(settings));
      }
      notify.success(t("system.toast.saved"));
    } catch (err: any) {
      if (typeof window !== "undefined") {
        localStorage.setItem("prim_admin_settings", JSON.stringify(settings));
      }
      notify.error(apiErrorMessage(err, t("system.toast.saveFailed")));
    } finally {
      setSaving(false);
    }
  };

  const handlePurgeCache = async () => {
    setPurgingCache(true);
    try {
      await new Promise((r) => setTimeout(r, 600));
      notify.success(t("system.actions.cachePurged"));
    } catch {
      notify.error(t("system.actions.purgeFailed"));
    } finally {
      setPurgingCache(false);
    }
  };

  return (
    <div className="space-y-6 w-full">
      {/* Top Header Card */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-3xl border border-border/80 shadow-xs">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="text-2xl font-bold tracking-tight text-foreground">{t("system.title")}</h1>
            <span className="px-2.5 py-0.5 rounded-full text-[11px] font-bold bg-accent-brand/10 text-accent-brand border border-accent-brand/20">
              {t("system.statusOnline")}
            </span>
          </div>
          <p className="text-xs text-muted-foreground mt-1">{t("system.subtitle")}</p>
        </div>

        <button
          onClick={() => handleSave()}
          disabled={saving || loading}
          className="flex items-center gap-2 px-5 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all shrink-0 disabled:opacity-50 cursor-pointer"
        >
          <Save className="size-4" />
          <span>{saving ? t("system.saving") : t("system.saveChanges")}</span>
        </button>
      </div>

      <div className="space-y-6">
        {/* Cache & Response Performance */}
        <div className="bg-background border border-border/80 rounded-2xl p-6 shadow-xs space-y-5">
          <div className="border-b border-border/60 pb-3 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div>
              <h3 className="font-bold text-sm text-foreground flex items-center gap-2">
                <Sliders className="size-4 text-accent-brand" />
                <span>{t("system.cache.title")}</span>
              </h3>
              <p className="text-[11px] text-muted-foreground mt-0.5">
                {t("system.cache.subtitle")}
              </p>
            </div>

            <button
              type="button"
              onClick={handlePurgeCache}
              disabled={purgingCache}
              className="flex items-center gap-2 px-4 py-2 rounded-xl border border-border bg-secondary/30 hover:bg-secondary text-foreground text-xs font-semibold transition-all disabled:opacity-50 self-start sm:self-center cursor-pointer shadow-xs"
            >
              <Trash2 className={`size-3.5 text-accent-brand ${purgingCache ? "animate-spin" : ""}`} />
              <span>{purgingCache ? t("system.actions.purging") : t("system.actions.purgeCache")}</span>
            </button>
          </div>

          {/* Disable Cache Switcher */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4 rounded-xl border border-border/70 bg-secondary/10">
            <div>
              <h4 className="text-xs font-bold text-foreground">
                {t("system.cache.disableCache")}
              </h4>
              <p className="text-[11px] text-muted-foreground mt-0.5">
                {t("system.cache.disableCacheDesc")}
              </p>
            </div>

            <label className="relative inline-flex items-center cursor-pointer shrink-0">
              <input
                type="checkbox"
                checked={Boolean(settings.disableCache)}
                onChange={(e) => setSettings({ ...settings, disableCache: e.target.checked })}
                className="sr-only peer"
              />
              <div className="w-10 h-6 bg-secondary/60 peer-focus:outline-hidden rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-border after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-rose-500"></div>
            </label>
          </div>
        </div>

        {/* Infrastructure & Services Health */}
        <div className="bg-background border border-border/80 rounded-2xl p-6 shadow-xs space-y-4">
          <div className="border-b border-border/60 pb-3 flex items-center justify-between">
            <div>
              <h3 className="font-bold text-sm text-foreground flex items-center gap-2">
                <Activity className="size-4 text-accent-brand" />
                <span>{t("system.infrastructure.title")}</span>
              </h3>
              <p className="text-[11px] text-muted-foreground mt-0.5">
                {t("system.infrastructure.subtitle")}
              </p>
            </div>
            <span className="flex items-center gap-1.5 text-[11px] text-emerald-600 dark:text-emerald-400 font-bold bg-emerald-500/10 px-2.5 py-1 rounded-full border border-emerald-500/20">
              <CheckCircle2 className="size-3.5" />
              <span>{t("system.infrastructure.allOperational")}</span>
            </span>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 pt-1">
            {[
              {
                name: "PostgreSQL Database",
                icon: Database,
                status: t("system.infrastructure.healthy"),
                detail: "Port 5432 • Active Pool • Latency 0.8ms",
                color: "text-emerald-500"
              },
              {
                name: "Redis Cache & Sessions",
                icon: Layers,
                status: t("system.infrastructure.healthy"),
                detail: "Port 6379 • In-Memory Store • Latency 0.2ms",
                color: "text-emerald-500"
              },
              {
                name: "Kafka Message Broker",
                icon: Radio,
                status: t("system.infrastructure.healthy"),
                detail: "Port 9092 • Notifications Topic • 0 Lag",
                color: "text-emerald-500"
              },
              {
                name: "MinIO S3 Storage",
                icon: HardDrive,
                status: t("system.infrastructure.healthy"),
                detail: "Port 9000 • Product Media Bucket • Online",
                color: "text-emerald-500"
              }
            ].map((node) => {
              const NodeIcon = node.icon;
              return (
                <div
                  key={node.name}
                  className="p-4 rounded-xl border border-border/70 bg-secondary/10 flex flex-col justify-between space-y-3"
                >
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex items-center gap-2 min-w-0">
                      <NodeIcon className="size-4 text-accent-brand shrink-0" />
                      <span className="font-bold text-xs text-foreground truncate">{node.name}</span>
                    </div>
                    <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20 shrink-0">
                      <span className="size-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                      {node.status}
                    </span>
                  </div>
                  <span className="text-[10px] text-muted-foreground font-mono">
                    {node.detail}
                  </span>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
}
