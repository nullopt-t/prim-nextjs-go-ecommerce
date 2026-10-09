"use client";

import { useEffect, useState } from "react";
import { useAuthContext } from "@/context/AuthContext";
import { userService } from "@/services/user";
import { toast } from "sonner";
import { useTranslations } from "next-intl";

interface Session {
  id: string;
  userAgent?: string;
  ipAddress?: string;
  createdAt?: string;
  lastSeenAt?: string;
  isCurrent?: boolean;
}

function SessionsSection() {
  const t = useTranslations("settings.sessions");
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);
      try {
        const res = await userService.getSessions();
        if (cancelled) return;
        const raw: Session[] = Array.isArray(res) ? res : res?.data ?? [];
        setSessions(raw);
      } catch (err: unknown) {
        if (!cancelled) setError(err instanceof Error ? err.message : t("error"));
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    load();
    return () => {
      cancelled = true;
    };
  }, [t]);

  const handleDelete = async (id: string) => {
    try {
      await userService.deleteSession(id);
      setSessions((prev) => prev.filter((s) => s.id !== id));
      toast.success(t("removedSuccess"));
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : t("error"));
    }
  };

  return (
    <div className="flex flex-col gap-3">
      <h3 className="text-sm font-semibold text-foreground">{t("title")}</h3>

      {loading ? (
        <div className="flex flex-col gap-2">
          <div className="animate-pulse bg-secondary/50 rounded-xl h-12" />
          <div className="animate-pulse bg-secondary/50 rounded-xl h-12" />
        </div>
      ) : error ? (
        <p className="text-destructive text-xs">{error}</p>
      ) : sessions.length === 0 ? (
        <p className="text-xs text-muted-foreground">{t("empty")}</p>
      ) : (
        <div className="flex flex-col gap-2">
          {sessions.map((session) => (
            <div
              key={session.id}
              className="bg-card border border-border/80 rounded-xl px-4 py-3 flex items-center justify-between gap-3 shadow-2xs"
            >
              <div className="flex flex-col min-w-0">
                <span className="text-xs sm:text-sm text-foreground font-medium truncate">
                  {session.userAgent
                    ? session.userAgent.slice(0, 60) + (session.userAgent.length > 60 ? "…" : "")
                    : t("unknownDevice")}
                  {session.isCurrent && (
                    <span className="ms-2 inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-semibold bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
                      {t("current")}
                    </span>
                  )}
                </span>
                {session.ipAddress && (
                  <span className="text-[11px] font-mono text-muted-foreground">{session.ipAddress}</span>
                )}
              </div>
              <button
                type="button"
                onClick={() => handleDelete(session.id)}
                disabled={session.isCurrent}
                aria-label="Remove session"
                className="text-xs font-medium text-destructive hover:underline transition-colors disabled:opacity-30 disabled:cursor-not-allowed shrink-0"
              >
                {t("remove")}
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export default function PersonalInfoForm() {
  const { user } = useAuthContext();
  const t = useTranslations("settings.personalInfo");

  return (
    <div className="flex flex-col gap-8 max-w-2xl">
      <form className="flex flex-col gap-5" onSubmit={(e) => e.preventDefault()}>
        {/* Notice */}
        <div className="rounded-xl border border-border/70 bg-secondary/30 px-4 py-3 text-xs text-muted-foreground leading-relaxed">
          {t("notice")}
        </div>

        {/* Display Name */}
        <div className="flex flex-col gap-1.5">
          <label htmlFor="display-name" className="text-xs font-medium text-foreground">
            {t("displayName")}
          </label>
          <input
            id="display-name"
            type="text"
            defaultValue={user?.name || ""}
            readOnly
            disabled
            placeholder={t("displayName")}
            className="h-10 w-full rounded-xl border border-border bg-secondary/30 px-3.5 text-xs sm:text-sm text-foreground placeholder:text-muted-foreground disabled:opacity-75 disabled:cursor-not-allowed focus:outline-none transition"
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
            value={user?.phone || ""}
            readOnly
            disabled
            placeholder={t("notSet")}
            className="h-10 w-full rounded-xl border border-border bg-secondary/30 px-3.5 text-xs sm:text-sm text-foreground placeholder:text-muted-foreground disabled:opacity-75 disabled:cursor-not-allowed focus:outline-none transition"
          />
          <p className="text-[11px] text-muted-foreground">
            {t("phoneNotice")}
          </p>
        </div>
      </form>

      {/* Divider */}
      <div className="border-t border-border/70" />

      {/* Active Sessions */}
      <SessionsSection />
    </div>
  );
}
