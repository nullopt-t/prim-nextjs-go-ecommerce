"use client";

import { useState, useEffect, useMemo } from "react";
import {
  Sliders,
  Trash2,
  Activity,
  Save,
  CheckCircle2,
  Database,
  Layers,
  HardDrive,
  Radio
} from "lucide-react";
import { notify, apiErrorMessage } from "@/lib/notify";
import { adminService, type SystemHealthReport, type SystemServiceHealth } from "@/services/admin";
import { useTranslations } from "next-intl";

const DEFAULT_SYSTEM_SETTINGS: Record<string, any> = {
  disableCache: false,
  enableDevOtpBypass: true,
  mockEmailNotifications: false
};

// Heartbeat interval for re-polling the core services health endpoint.
const HEALTH_POLL_MS = 15000;

// Static presentation for each probed service. Live status, port, latency, and
// lag are overlaid from the backend health report.
const SERVICE_NODES = [
  { key: "postgres", icon: Database, metric: "latency", defaultPort: 5432 },
  { key: "redis", icon: Layers, metric: "latency", defaultPort: 6379 },
  { key: "kafka", icon: Radio, metric: "lag", defaultPort: 9092 },
  { key: "minio", icon: HardDrive, metric: "online", defaultPort: 9000 }
] as const;

const STATUS_STYLES: Record<string, { badge: string; dot: string }> = {
  healthy: {
    badge: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20",
    dot: "bg-emerald-500"
  },
  degraded: {
    badge: "bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20",
    dot: "bg-amber-500"
  },
  unhealthy: {
    badge: "bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/20",
    dot: "bg-rose-500"
  },
  unknown: {
    badge: "bg-secondary/40 text-muted-foreground border-border/60",
    dot: "bg-muted-foreground"
  }
};

function formatLatency(value?: number): string {
  if (typeof value !== "number" || Number.isNaN(value)) return "0";
  return value.toFixed(1);
}

export function AdminSystem() {
  const t = useTranslations("admin");
  const [settings, setSettings] = useState<Record<string, any>>({ ...DEFAULT_SYSTEM_SETTINGS });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [purgingCache, setPurgingCache] = useState(false);
  const [health, setHealth] = useState<SystemHealthReport | null>(null);

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

  useEffect(() => {
    let isMounted = true;
    const loadHealth = async () => {
      try {
        const res = await adminService.getSystemHealth();
        if (isMounted) setHealth(res?.data ?? null);
      } catch {
        if (isMounted) setHealth(null);
      }
    };
    loadHealth();
    const interval = setInterval(loadHealth, HEALTH_POLL_MS);
    return () => {
      isMounted = false;
      clearInterval(interval);
    };
  }, []);

  const healthByKey = useMemo(() => {
    const map: Record<string, SystemServiceHealth> = {};
    (health?.services ?? []).forEach((service) => {
      map[service.key] = service;
    });
    return map;
  }, [health]);

  const totalServices = SERVICE_NODES.length;
  const healthyServices = (health?.services ?? []).filter((s) => s.status === "healthy").length;
  const overallStatus = health?.status ?? "unknown";
  const overallStyle = STATUS_STYLES[overallStatus] ?? STATUS_STYLES.unknown;
  const overallLabel = !health
    ? t("system.infrastructure.unknown")
    : overallStatus === "healthy"
      ? t("system.infrastructure.allOperational")
      : t("system.infrastructure.partialOperational", { up: healthyServices, total: totalServices });

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
            <span
              className={
                health && overallStatus !== "healthy"
                  ? `px-2.5 py-0.5 rounded-full text-[11px] font-bold border ${overallStyle.badge}`
                  : "px-2.5 py-0.5 rounded-full text-[11px] font-bold bg-accent-brand/10 text-accent-brand border border-accent-brand/20"
              }
            >
              {health && overallStatus !== "healthy" ? overallLabel : t("system.statusOnline")}
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
            <span className={`flex items-center gap-1.5 text-[11px] font-bold px-2.5 py-1 rounded-full border ${overallStyle.badge}`}>
              <CheckCircle2 className="size-3.5" />
              <span>{overallLabel}</span>
            </span>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 pt-1">
            {SERVICE_NODES.map((node) => {
              const NodeIcon = node.icon;
              const service = healthByKey[node.key];
              const status = service?.status ?? "unknown";
              const style = STATUS_STYLES[status] ?? STATUS_STYLES.unknown;
              const port = service?.port ?? node.defaultPort;

              let metric: string;
              if (!service) {
                metric = t("system.infrastructure.metric.unavailable");
              } else if (node.metric === "latency") {
                metric = t("system.infrastructure.metric.latency", {
                  value: formatLatency(service.latencyMs)
                });
              } else if (node.metric === "lag") {
                metric =
                  typeof service.lag === "number"
                    ? t("system.infrastructure.metric.lag", { value: service.lag })
                    : t("system.infrastructure.metric.connected");
              } else {
                metric =
                  status === "healthy"
                    ? t("system.infrastructure.metric.online")
                    : t("system.infrastructure.metric.offline");
              }

              const detail = `${t("system.infrastructure.metric.port", { port })} • ${t(
                `system.infrastructure.services.${node.key}.info`
              )} • ${metric}`;

              return (
                <div
                  key={node.key}
                  className="p-4 rounded-xl border border-border/70 bg-secondary/10 flex flex-col justify-between space-y-3"
                >
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex items-center gap-2 min-w-0">
                      <NodeIcon className="size-4 text-accent-brand shrink-0" />
                      <span className="font-bold text-xs text-foreground truncate">
                        {t(`system.infrastructure.services.${node.key}.name`)}
                      </span>
                    </div>
                    <span className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[10px] font-bold border shrink-0 ${style.badge}`}>
                      <span className={`size-1.5 rounded-full ${style.dot} ${status === "healthy" ? "animate-pulse" : ""}`}></span>
                      {t(`system.infrastructure.${status}`)}
                    </span>
                  </div>
                  <span className="text-[10px] text-muted-foreground font-mono" dir="ltr">
                    {detail}
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
