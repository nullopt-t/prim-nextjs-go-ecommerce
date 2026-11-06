"use client";

import { useState, useEffect, useMemo } from "react";
import { adminService, type SystemHealthReport, type SystemServiceHealth } from "@/modules/admin/api/adminService";
import { useTranslations } from "next-intl";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill, type StatusTone } from "@/shared/ui/statusPill";

// Short-poll interval for the core services health endpoint. Kept short so the
// panel feels live; a single in-flight guard prevents request pile-up when a
// report is slow (probes can take a few seconds under load).
const HEALTH_POLL_MS = 5000;

// Number of latency samples retained per service for the trend sparkline.
const HISTORY_LIMIT = 24;

type ServiceStatusKey = "healthy" | "degraded" | "unhealthy" | "unknown";

// Probed services, in display order. Live status, latency, and resource
// readings are overlaid from the backend health report.
const SERVICE_NODES = ["postgres", "redis", "kafka", "minio"] as const;

const STATUS_TONES: Record<ServiceStatusKey, StatusTone> = {
  healthy: "success",
  degraded: "warning",
  unhealthy: "danger",
  unknown: "neutral",
};

const SPARK_TONES: Record<ServiceStatusKey, string> = {
  healthy: "text-emerald-500",
  degraded: "text-amber-500",
  unhealthy: "text-rose-500",
  unknown: "text-muted-foreground",
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
  const overallTone = STATUS_TONES[overallStatus] ?? "neutral";
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
    <div className="w-full space-y-5">
      <PageHeader title={t("system.title")} />

      <section className="space-y-4 rounded-lg border border-border bg-background p-4 sm:p-5">
        {refreshing ? (
          <span className="pointer-events-none block h-0.5 overflow-hidden" aria-hidden="true">
            <span className="prim-sweep block h-full w-1/3 rounded-full bg-primary/40" />
          </span>
        ) : null}

        <div className="flex flex-wrap items-start justify-between gap-3 border-b border-border pb-3">
          <div>
            <h2 className="text-sm font-semibold text-foreground">{t("system.infrastructure.title")}</h2>
            <p className="mt-0.5 text-xs text-muted-foreground">{t("system.infrastructure.subtitle")}</p>
          </div>

          <div className="flex shrink-0 flex-col items-end gap-1.5">
            <StatusPill tone={overallTone}>{overallLabel}</StatusPill>
            {health ? (
              <span className={`flex items-center gap-1.5 font-mono text-xs ${isStale ? "text-amber-700 dark:text-amber-400" : "text-muted-foreground"}`}>
                <span className={`size-1.5 rounded-full ${freshnessDot}`} />
                <span>{freshnessLabel}</span>
              </span>
            ) : null}
          </div>
        </div>

        <div className="grid grid-cols-1 gap-3 pt-1 sm:grid-cols-2 lg:grid-cols-4">
          {!healthLoaded
            ? SERVICE_NODES.map((key) => (
                <div key={key} className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
                  <div className="flex items-center gap-2.5">
                    <span className="prim-shimmer size-8 shrink-0 rounded-md bg-muted" />
                    <span className="prim-shimmer h-3 w-2/3 rounded bg-muted" />
                  </div>
                  <span className="prim-shimmer h-2.5 w-full rounded bg-muted" />
                  <span className="prim-shimmer h-2.5 w-4/5 rounded bg-muted" />
                  <span className="prim-shimmer h-6 w-full rounded bg-muted" />
                </div>
              ))
            : SERVICE_NODES.map((key) => {
                const service = healthByKey[key];
                const status = (service?.status ?? "unknown") as ServiceStatusKey;
                const resources = service?.resources ?? [];
                const latencyLabel = service
                  ? t("system.infrastructure.metric.latency", { value: formatLatency(service.latencyMs) })
                  : t("system.infrastructure.metric.unavailable");

                return (
                  <div key={key} className="flex flex-col justify-between gap-3 rounded-lg border border-border bg-card p-4">
                    <div className="flex items-center justify-between gap-2">
                      <span className="truncate text-sm font-medium text-foreground">
                        {t(`system.infrastructure.services.${key}.name`)}
                      </span>
                      <StatusPill tone={STATUS_TONES[status] ?? "neutral"}>
                        {t(`system.infrastructure.${status}`)}
                      </StatusPill>
                    </div>

                    <dl className="space-y-1.5">
                      {resources.length > 0 ? (
                        resources.map((resource) => (
                          <div key={resource.key} className="flex items-baseline justify-between gap-3 text-xs">
                            <dt className="truncate text-muted-foreground">
                              {t(`system.infrastructure.resources.${resource.key}`)}
                            </dt>
                            <dd className="font-mono font-medium text-foreground tabular-nums">{resource.value}</dd>
                          </div>
                        ))
                      ) : (
                        <div className="text-xs text-muted-foreground">
                          {t("system.infrastructure.metric.unavailable")}
                        </div>
                      )}
                    </dl>

                    <div className="space-y-2 border-t border-border pt-2">
                      <span className="block font-mono text-xs text-muted-foreground">{latencyLabel}</span>
                      <Sparkline values={history[key] ?? []} className={SPARK_TONES[status] ?? SPARK_TONES.unknown} />
                    </div>
                  </div>
                );
              })}
        </div>
      </section>
    </div>
  );
}
