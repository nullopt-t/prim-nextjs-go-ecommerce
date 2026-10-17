"use client";

import { useTranslations } from "next-intl";
import { Link, usePathname } from "@/i18n/navigation";
import {
  User,
  Sliders,
  ShieldAlert,
  MonitorSmartphone,
  ChevronRight,
  ChevronLeft,
  ArrowLeft,
  ArrowRight,
} from "lucide-react";
import { useLocale } from "next-intl";

export const settingsSubPages = [
  {
    id: "profile",
    path: "/user/settings/profile",
    labelKey: "personalInfo.title",
    descKey: "personalInfo.subtitle",
    icon: User,
    isDanger: false,
  },
  {
    id: "preferences",
    path: "/user/settings/preferences",
    labelKey: "preferences.title",
    descKey: "preferences.subtitle",
    icon: Sliders,
    isDanger: false,
  },
  {
    id: "sessions",
    path: "/user/settings/sessions",
    labelKey: "sessions.title",
    descKey: "sessions.subtitle",
    icon: MonitorSmartphone,
    isDanger: false,
  },
  {
    id: "danger",
    path: "/user/settings/danger",
    labelKey: "danger.title",
    descKey: "danger.subtitle",
    icon: ShieldAlert,
    isDanger: true,
  },
] as const;

export type SettingsTabId = (typeof settingsSubPages)[number]["id"];

interface SettingsSubNavProps {
  activeId?: SettingsTabId;
}

export function SettingsSubNav({ activeId }: SettingsSubNavProps) {
  const t = useTranslations("settings");
  const pathname = usePathname();

  return (
    <div className="w-full mb-6">
      {/* ── Desktop Tab Bar (sm: and up) ── */}
      <div className="hidden sm:block border-b border-border/70">
        <nav aria-label={t("subNav")} className="flex items-center gap-2 overflow-x-auto scrollbar-none -mb-px">
          {settingsSubPages.map((item) => {
            const isActive = activeId
              ? activeId === item.id
              : pathname === item.path || pathname.startsWith(item.path);

            const Icon = item.icon;

            return (
              <Link
                key={item.id}
                href={item.path}
                className={`inline-flex items-center gap-2 py-2.5 px-3.5 text-xs sm:text-sm font-medium border-b-2 whitespace-nowrap transition-colors duration-150 cursor-pointer ${
                  isActive
                    ? item.isDanger
                      ? "border-destructive text-destructive font-semibold"
                      : "border-foreground text-foreground font-semibold"
                    : item.isDanger
                    ? "border-transparent text-muted-foreground hover:text-destructive hover:border-destructive/40"
                    : "border-transparent text-muted-foreground hover:text-foreground hover:border-border"
                }`}
              >
                <Icon className="size-4" />
                <span>{t(item.labelKey as any)}</span>
              </Link>
            );
          })}
        </nav>
      </div>
    </div>
  );
}

interface SettingsShellProps {
  children: React.ReactNode;
  activeId?: SettingsTabId;
  isSubPage?: boolean;
}

export default function SettingsShell({ children, isSubPage = false }: SettingsShellProps) {
  const t = useTranslations("settings");
  const locale = useLocale();
  const isRtl = locale === "ar";
  const BackIcon = isRtl ? ArrowRight : ArrowLeft;

  return (
    <div className="w-full flex flex-col pb-12">
      {/* Mobile-only "Back to Settings" strip when inside a sub-page */}
      {isSubPage && (
        <div className="sm:hidden pb-3 border-b border-border/60 mb-4">
          <Link
            href="/user/settings"
            className="inline-flex items-center gap-1.5 text-xs font-medium text-accent-brand hover:underline cursor-pointer"
          >
            <BackIcon className="size-3.5" />
            <span>{t("backToSettings")}</span>
          </Link>
        </div>
      )}

      {/* ── Sub Page Content ── */}
      <div className="w-full">{children}</div>
    </div>
  );
}

/**
 * Mobile-first Settings hub listing: Amazon-style links list on small screens,
 * redirecting to the specific settings sub-pages.
 */
export function SettingsLinksMenu() {
  const t = useTranslations("settings");
  const locale = useLocale();
  const isRtl = locale === "ar";
  const Chevron = isRtl ? ChevronLeft : ChevronRight;

  return (
    <div className="w-full flex flex-col divide-y divide-border/70 border border-border/80 rounded-2xl bg-card overflow-hidden shadow-2xs">
      {settingsSubPages.map((item) => {
        const Icon = item.icon;

        return (
          <Link
            key={item.id}
            href={item.path}
            className={`group p-4 flex items-center justify-between gap-4 transition-colors cursor-pointer hover:bg-secondary/40 ${
              item.isDanger ? "hover:bg-destructive/5" : ""
            }`}
          >
            <div className="flex items-center gap-3.5 min-w-0">
              <div
                className={`size-10 rounded-xl flex items-center justify-center shrink-0 border transition-colors ${
                  item.isDanger
                    ? "bg-destructive/10 text-destructive border-destructive/20 group-hover:bg-destructive/15"
                    : "bg-secondary text-foreground border-border/80 group-hover:bg-secondary/80"
                }`}
              >
                <Icon className="size-5" />
              </div>
              <div className="flex flex-col min-w-0">
                <span
                  className={`text-sm font-semibold truncate ${
                    item.isDanger ? "text-destructive" : "text-foreground group-hover:text-accent-brand"
                  }`}
                >
                  {t(item.labelKey as any)}
                </span>
                <span className="text-xs text-muted-foreground line-clamp-1 mt-0.5">
                  {t(item.descKey as any)}
                </span>
              </div>
            </div>

            <Chevron
              className={`size-4.5 text-muted-foreground/70 shrink-0 transition-transform ${
                item.isDanger
                  ? "group-hover:text-destructive"
                  : "group-hover:text-accent-brand"
              }`}
            />
          </Link>
        );
      })}
    </div>
  );
}
