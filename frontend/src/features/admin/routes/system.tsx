"use client";

import { useState, useEffect, useMemo } from "react";
import {
  Activity,
  CheckCircle2,
  AlertTriangle,
  Database,
  Layers,
  HardDrive,
  Radio
} from "lucide-react";
import { adminService, type SystemHealthReport, type SystemServiceHealth } from "@/services/admin";
import { useTranslations } from "next-intl";

// Short-poll interval for the core services health endpoint. Kept short so the
// panel feels live; a single in-flight guard prevents request pile-up when a
// report is slow (probes can take a few seconds under load).
const HEALTH_POLL_MS = 5000;

// Number of latency samples retained per service for the trend sparkline.
const HISTORY_LIMIT = 24;

type ServiceStatusKey = "healthy" | "degraded" | "unhealthy" | "unknown";

// Static presentation for each probed service. Live status, latency, and
// resource readings are overlaid from the backend health report.
const SERVICE_NODES = [
  { key: "postgres", icon: Database },
  { key: "redis", icon: Layers },
  { key: "kafka", icon: Radio },
  { key: "minio", icon: HardDrive }
] as const;

const STATUS_STYLES: Record<ServiceStatusKey, { badge: string; dot: string; tile: string; spark: string }> = {
  healthy: {
    badge: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20",
    dot: "bg-emerald-500",
    tile: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400",
    spark: "text-emerald-500"
  },
  degraded: {
    badge: "bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20",
    dot: "bg-amber-500",
    tile: "bg-amber-500/10 text-amber-600 dark:text-amber-400",
    spark: "text-amber-500"
  },
  unhealthy: {
    badge: "bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/20",
    dot: "bg-rose-500",
    tile: "bg-rose-500/10 text-rose-600 dark:text-rose-400",
    spark: "text-rose-500"
  },
  unknown: {
    badge: "bg-secondary/40 text-muted-foreground border-border/60",
    dot: "bg-muted-foreground",
    tile: "bg-secondary/50 text-muted-foreground",
    spark: "text-muted-foreground"
  }
};

function formatLatency(value?: number): string {
  if (typeof value !== "number" || Number.isNaN(value)) return "0";
  return value.toFixed(1);
}

// Sparkline renders a lightweight latency trend from the rolling sample buffer.
// It preserves the request order (left → right) in both LTR and RTL layouts.
function Sparkline({ values, className }: { values: number[]; className?: string }) {
  const data = values.slice(-HISTORY_LIMIT);
  if (data.length < 2) {
    return <div className="h-6 w-full" aria-hidden="true" />;
  }

  const width = 100;
  const height = 24;
  const min = Math.min(...data);
  const max = Math.max(...data);
  const span = max - min || 1;
  const step = width / (data.length - 1);
  const points = data
    .map((value, index) => {
      const x = index * step;
      const y = height - ((value - min) / span) * height;
      return `${x.toFixed(2)},${y.toFixed(2)}`;
    })
    .join(" ");
  const area = `0,${height} ${points} ${width},${height}`;

  return (
    <svg
      viewBox={`0 0 ${width} ${height}`}
      preserveAspectRatio="none"
      aria-hidden="true"
      className={`h-6 w-full ${className ?? ""}`}
    >
      <polyline points={area} fill="currentColor" fillOpacity="0.08" stroke="none" />
      <polyline
        points={points}
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
        vectorEffect="non-scaling-stroke"
      />
    </svg>
  );
}

export function AdminSystem() {
  const t = useTranslations("admin");
  const [health, setHealth] = useState<SystemHealthReport | null>(null);
  const [healthLoaded, setHealthLoaded] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [history, setHistory] = useState<Record<string, number[]>>({});
  const [now, setNow] = useState(() => Date.now());

  // Short polling: fetch the health report on a fixed interval, skipping a tick
  // while the previous request is still in flight so a slow report cannot pile
  // up.
  useEffect(() => {
    let isMounted = true;
    let inFlight = false;

    const loadHealth = async () => {
      if (inFlight) return;
      inFlight = true;
      if (isMounted) setRefreshing(true);
      try {
        const res = await adminService.getSystemHealth();
        const report = res?.data ?? null;
        if (!isMounted || !report) return;
        setHealth(report);
        setHistory((prev) => {
          const next = { ...prev };
          report.services.forEach((service) => {
            const series = [...(next[service.key] ?? []), service.latencyMs];
            next[service.key] = series.slice(-HISTORY_LIMIT);
          });
          return next;
        });
      } catch {
        // Keep the last known report; the freshness label signals staleness.
      } finally {
        inFlight = false;
        if (isMounted) {
          setRefreshing(false);
          setHealthLoaded(true);
        }
      }
    };

    loadHealth();
    const poll = setInterval(loadHealth, HEALTH_POLL_MS);
    return () => {
      isMounted = false;
      clearInterval(poll);
    };
  }, []);

  // Tick the "updated Xs ago" label every second so it tracks the live stream.
  useEffect(() => {
    const tick = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(tick);
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
  const overallStatus = (health?.status ?? "unknown") as ServiceStatusKey;
  const overallStyle = STATUS_STYLES[overallStatus] ?? STATUS_STYLES.unknown;
  const overallLabel = !health
    ? healthLoaded
      ? t("system.infrastructure.metric.unavailable")
      : t("system.infrastructure.checking")
    : overallStatus === "healthy"
      ? t("system.infrastructure.allOperational")
      : t("system.infrastructure.partialOperational", { up: healthyServices, total: totalServices });

  const checkedAtMs = health ? Date.parse(health.checkedAt) : NaN;
  const secondsAgo = Number.isFinite(checkedAtMs) ? Math.max(0, Math.round((now - checkedAtMs) / 1000)) : null;
  const isStale = secondsAgo !== null && secondsAgo * 1000 > HEALTH_POLL_MS * 2;
  const freshnessLabel = health
    ? t("system.infrastructure.updated", { seconds: secondsAgo ?? 0 })
    : healthLoaded
      ? t("system.infrastructure.metric.unavailable")
      : t("system.infrastructure.checking");
  const freshnessDot = !health
    ? healthLoaded
      ? "bg-rose-500"
      : "bg-muted-foreground"
    : isStale
      ? "bg-amber-500"
      : "bg-emerald-500";

  return (
    <div className="space-y-6 w-full">
      {/* Top Header Card */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-3xl border border-border/80 shadow-xs">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="text-2xl font-bold tracking-tight text-foreground">{t("system.title")}</h1>
            <span
              className={
                !health || overallStatus !== "healthy"
                  ? `px-2.5 py-0.5 rounded-full text-[11px] font-bold border ${overallStyle.badge}`
                  : "px-2.5 py-0.5 rounded-full text-[11px] font-bold bg-accent-brand/10 text-accent-brand border border-accent-brand/20"
              }
            >
              {!health || overallStatus !== "healthy" ? overallLabel : t("system.statusOnline")}
            </span>
          </div>
          <p className="text-xs text-muted-foreground mt-1">{t("system.subtitle")}</p>
        </div>
      </div>

      <div className="space-y-6">
        {/* Infrastructure & Services Health */}
        <div className="relative overflow-hidden bg-background border border-border/80 rounded-2xl p-6 shadow-xs space-y-4">
          {refreshing && (
            <span className="pointer-events-none absolute inset-x-0 top-0 h-[2px] overflow-hidden" aria-hidden="true">
              <span className="prim-sweep block h-full w-1/3 rounded-full bg-accent-brand/50" />
            </span>
          )}

          <div className="border-b border-border/60 pb-3 flex items-start justify-between gap-3">
            <div>
              <h3 className="font-bold text-sm text-foreground flex items-center gap-2">
                <Activity className="size-4 text-accent-brand" />
                <span>{t("system.infrastructure.title")}</span>
              </h3>
              <p className="text-[11px] text-muted-foreground mt-0.5">
                {t("system.infrastructure.subtitle")}
              </p>
            </div>

            <div className="flex flex-col items-end gap-1.5 shrink-0">
              <span className={`flex items-center gap-1.5 text-[11px] font-bold px-2.5 py-1 rounded-full border ${overallStyle.badge}`}>
                {overallStatus === "healthy" ? (
                  <CheckCircle2 className="size-3.5" />
                ) : overallStatus === "unknown" ? (
                  <Activity className="size-3.5" />
                ) : (
                  <AlertTriangle className="size-3.5" />
                )}
                <span>{overallLabel}</span>
              </span>
              {health && (
                <span className={`flex items-center gap-1.5 text-[10px] font-mono ${isStale ? "text-amber-600 dark:text-amber-400" : "text-muted-foreground"}`}>
                  <span className="relative flex size-1.5">
                    <span className={`prim-heartbeat absolute inline-flex size-full rounded-full ${freshnessDot}`} />
                    <span className={`relative inline-flex size-1.5 rounded-full ${freshnessDot}`} />
                  </span>
                  <span>{freshnessLabel}</span>
                </span>
              )}
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 pt-1">
            {!healthLoaded
              ? SERVICE_NODES.map((node) => (
                  <div
                    key={node.key}
                    className="p-4 rounded-xl border border-border/70 bg-card flex flex-col gap-3"
                  >
                    <div className="flex items-center gap-2.5">
                      <span className="prim-shimmer size-8 shrink-0 rounded-md bg-muted" />
                      <span className="prim-shimmer h-3 w-2/3 rounded bg-muted" />
                    </div>
                    <span className="prim-shimmer h-2.5 w-full rounded bg-muted" />
                    <span className="prim-shimmer h-2.5 w-4/5 rounded bg-muted" />
                    <span className="prim-shimmer h-6 w-full rounded bg-muted" />
                  </div>
                ))
              : SERVICE_NODES.map((node) => {
                  const NodeIcon = node.icon;
                  const service = healthByKey[node.key];
                  const status = (service?.status ?? "unknown") as ServiceStatusKey;
                  const style = STATUS_STYLES[status] ?? STATUS_STYLES.unknown;
                  const resources = service?.resources ?? [];
                  const latencyLabel = service
                    ? t("system.infrastructure.metric.latency", { value: formatLatency(service.latencyMs) })
                    : t("system.infrastructure.metric.unavailable");

                  return (
                    <div
                      key={node.key}
                      className="group p-4 rounded-xl border border-border/70 bg-card flex flex-col justify-between gap-3 transition-all duration-200 hover:-translate-y-0.5 hover:border-border hover:shadow-sm"
                    >
                      <div className="flex items-center justify-between gap-2">
                        <div className="flex items-center gap-2.5 min-w-0">
                          <span className={`grid size-8 shrink-0 place-items-center rounded-md transition-colors ${style.tile}`}>
                            <NodeIcon className="size-4" />
                          </span>
                          <span className="font-bold text-xs text-foreground truncate">
                            {t(`system.infrastructure.services.${node.key}.name`)}
                          </span>
                        </div>
                        <span className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[10px] font-bold border shrink-0 ${style.badge}`}>
                          <span className="relative flex size-1.5">
                            <span className={`prim-heartbeat absolute inline-flex size-full rounded-full ${style.dot}`} />
                            <span className={`relative inline-flex size-1.5 rounded-full ${style.dot}`} />
                          </span>
                          {t(`system.infrastructure.${status}`)}
                        </span>
                      </div>

                      <div className="space-y-1.5">
                        {resources.length > 0 ? (
                          resources.map((resource) => (
                            <div
                              key={resource.key}
                              className="flex items-baseline justify-between gap-3 text-[11px]"
                            >
                              <span className="text-muted-foreground truncate">
                                {t(`system.infrastructure.resources.${resource.key}`)}
                              </span>
                              <span className="font-mono font-semibold text-foreground tabular-nums">
                                {resource.value}
                              </span>
                            </div>
                          ))
                        ) : (
                          <div className="text-[11px] text-muted-foreground">
                            {t("system.infrastructure.metric.unavailable")}
                          </div>
                        )}
                      </div>

                      <div className="space-y-2 border-t border-border/60 pt-2">
                        <span className="block text-[10px] text-muted-foreground font-mono">
                          {latencyLabel}
                        </span>
                        <Sparkline values={history[node.key] ?? []} className={style.spark} />
                      </div>
                    </div>
                  );
                })}
          </div>
        </div>
      </div>
    </div>
  );
}
