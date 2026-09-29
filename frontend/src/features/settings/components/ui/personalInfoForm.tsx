"use client";

import { useEffect, useState } from "react";
import { useAuthContext } from "@/context/AuthContext";
import { userService } from "@/services/user";
import { Lock, Monitor, Trash2 } from "lucide-react";
import { toast } from "sonner";

// ── Types ──────────────────────────────────────────────────────────────────────
interface Session {
  id: string;
  userAgent?: string;
  ipAddress?: string;
  createdAt?: string;
  lastSeenAt?: string;
  isCurrent?: boolean;
}

// ── Sessions section ───────────────────────────────────────────────────────────
function SessionsSection() {
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
      } catch (err: any) {
        if (!cancelled) setError(err?.message || "Failed to load sessions.");
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    load();
    return () => {
      cancelled = true;
    };
  }, []);

  const handleDelete = async (id: string) => {
    try {
      await userService.deleteSession(id);
      setSessions((prev) => prev.filter((s) => s.id !== id));
      toast.success("Session removed.");
    } catch (err: any) {
      toast.error(err?.message || "Failed to remove session.");
    }
  };

  return (
    <div className="flex flex-col gap-3">
      <h3 className="text-sm font-semibold text-foreground">Active Sessions</h3>

      {loading ? (
        <div className="flex flex-col gap-2">
          <div className="animate-pulse bg-secondary rounded h-12" />
          <div className="animate-pulse bg-secondary rounded h-12" />
        </div>
      ) : error ? (
        <p className="text-destructive text-sm">Error: {error}</p>
      ) : sessions.length === 0 ? (
        <p className="text-sm text-muted-foreground">No active sessions found.</p>
      ) : (
        <div className="flex flex-col gap-2">
          {sessions.map((session) => (
            <div
              key={session.id}
              className="bg-card border border-border rounded-lg px-4 py-3 flex items-center justify-between gap-3 shadow-sm"
            >
              <div className="flex items-center gap-3">
                <Monitor className="size-4 text-muted-foreground shrink-0" />
                <div className="flex flex-col">
                  <span className="text-sm text-foreground font-medium">
                    {session.userAgent
                      ? session.userAgent.slice(0, 60) + (session.userAgent.length > 60 ? "…" : "")
                      : "Unknown device"}
                    {session.isCurrent && (
                      <span className="ml-2 inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-accent-brand/10 text-accent-brand">
                        Current
                      </span>
                    )}
                  </span>
                  {session.ipAddress && (
                    <span className="text-xs text-muted-foreground">{session.ipAddress}</span>
                  )}
                </div>
              </div>
              <button
                type="button"
                onClick={() => handleDelete(session.id)}
                disabled={session.isCurrent}
                aria-label="Remove session"
                className="h-9 px-3 rounded-md text-sm font-medium text-destructive hover:bg-destructive/10 transition disabled:opacity-40 disabled:cursor-not-allowed"
              >
                <Trash2 className="size-4" />
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

// ── Main component ─────────────────────────────────────────────────────────────
export default function PersonalInfoForm() {
  const { user } = useAuthContext();

  return (
    <div className="flex flex-col gap-8 max-w-xl">
      <form className="flex flex-col gap-5" onSubmit={(e) => e.preventDefault()}>
        {/* Profile updates unavailable notice */}
        <div className="rounded-md border border-border bg-secondary/50 px-4 py-3 text-sm text-muted-foreground">
          Profile updates are not available yet.
        </div>

        {/* Display Name */}
        <div className="flex flex-col gap-1.5">
          <label htmlFor="display-name" className="text-sm font-medium text-foreground">
            Display Name
          </label>
          <input
            id="display-name"
            type="text"
            defaultValue={user?.name || ""}
            readOnly
            disabled
            placeholder="Your display name"
            className="h-10 w-full rounded-md border border-border bg-input-background px-3 text-sm text-foreground placeholder:text-muted-foreground disabled:opacity-60 disabled:cursor-not-allowed focus:outline-none transition"
          />
        </div>

        {/* Email (read-only) */}
        <div className="flex flex-col gap-1.5">
          <label htmlFor="email" className="text-sm font-medium text-foreground">
            Email Address
          </label>
          <div className="relative">
            <input
              id="email"
              type="email"
              value={user?.email || ""}
              readOnly
              disabled
              className="h-10 w-full rounded-md border border-border bg-input-background px-3 pr-10 text-sm text-foreground placeholder:text-muted-foreground disabled:opacity-60 disabled:cursor-not-allowed focus:outline-none transition"
            />
            <Lock className="absolute right-3 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground/60" />
          </div>
          <p className="text-xs text-muted-foreground">
            Email cannot be changed. Contact support to update.
          </p>
        </div>

        {/* Phone (read-only) */}
        <div className="flex flex-col gap-1.5">
          <label htmlFor="phone" className="text-sm font-medium text-foreground">
            Phone Number
          </label>
          <input
            id="phone"
            type="tel"
            value={user?.phone || ""}
            readOnly
            disabled
            placeholder="Not set"
            className="h-10 w-full rounded-md border border-border bg-input-background px-3 text-sm text-foreground placeholder:text-muted-foreground disabled:opacity-60 disabled:cursor-not-allowed focus:outline-none transition"
          />
          <p className="text-xs text-muted-foreground">
            Contact support to update your phone number.
          </p>
        </div>

        {/* Save — disabled */}
        <div className="flex justify-end">
          <button
            type="submit"
            disabled
            className="h-9 px-4 rounded-md bg-accent-brand text-white text-sm font-medium opacity-50 cursor-not-allowed transition shadow-sm"
          >
            Save Changes
          </button>
        </div>
      </form>

      {/* Divider */}
      <div className="border-t border-border" />

      {/* Active Sessions */}
      <SessionsSection />
    </div>
  );
}
