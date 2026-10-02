"use client";

import { useState, useRef, useEffect } from "react";
import {
  Bell,
  Check,
  CheckCheck,
  Package,
  ShieldCheck,
  Info,
  X,
  ExternalLink,
} from "lucide-react";
import { useNotifications } from "@/context/NotificationContext";
import { useLocale, useTranslations } from "next-intl";
import { useRouter } from "@/i18n/navigation";

export function NotificationsDropdown() {
  const {
    notifications,
    unreadCount,
    loading,
    markAsRead,
    markAllAsRead,
    deleteNotification,
  } = useNotifications();

  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const locale = useLocale();
  const router = useRouter();
  const t = useTranslations("common.header");

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
      document.addEventListener("keydown", handleKeyDown);
    }
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [isOpen]);

  const getCategoryIcon = (category: string) => {
    switch (category) {
      case "order":
        return <Package className="size-4 text-emerald-500" />;
      case "auth":
        return <ShieldCheck className="size-4 text-blue-500" />;
      default:
        return <Info className="size-4 text-amber-500" />;
    }
  };

  const handleNotificationClick = async (item: any) => {
    if (!item.isRead) {
      await markAsRead(item.id);
    }
    if (item.actionUrl) {
      setIsOpen(false);
      router.push(item.actionUrl);
    }
  };

  return (
    <div className="relative" ref={dropdownRef}>
      {/* Bell Trigger Button */}
      <button
        type="button"
        onClick={() => setIsOpen((prev) => !prev)}
        aria-label={t("notifications")}
        aria-expanded={isOpen}
        className={`size-9 sm:size-10 rounded-xl transition-all relative flex items-center justify-center cursor-pointer border shadow-2xs ${
          isOpen
            ? "border-accent-brand/60 bg-secondary/80 text-foreground"
            : "border-border/40 text-muted-foreground hover:text-accent-brand hover:bg-secondary/70 hover:border-accent-brand/40"
        }`}
      >
        <Bell className="size-4.5 sm:size-5" />
        {unreadCount > 0 && (
          <span className="absolute -top-1 ltr:-right-1 rtl:-left-1 min-w-[18px] h-4.5 px-1 bg-rose-500 text-white text-[10px] font-bold rounded-full flex items-center justify-center shadow-xs ring-2 ring-background animate-in zoom-in duration-150">
            {unreadCount > 99 ? "99+" : unreadCount}
          </span>
        )}
      </button>

      {/* Notifications Drawer / Dropdown */}
      {isOpen && (
        <div
          className={`absolute mt-2 w-80 sm:w-96 rounded-2xl bg-card border border-border/80 shadow-2xl z-50 overflow-hidden animate-in fade-in-0 zoom-in-95 duration-150 ${
            locale === "ar" ? "left-0 origin-top-left" : "right-0 origin-top-right"
          }`}
        >
          {/* Header */}
          <div className="px-4 py-3 border-b border-border/60 flex items-center justify-between bg-muted/30">
            <div className="flex items-center gap-2">
              <h3 className="font-semibold text-foreground text-sm">
                {t("notifications")}
              </h3>
              {unreadCount > 0 && (
                <span className="text-[11px] font-medium px-2 py-0.5 rounded-full bg-accent-brand/10 text-accent-brand">
                  {unreadCount} {locale === "ar" ? "جديد" : "new"}
                </span>
              )}
            </div>
            {unreadCount > 0 && (
              <button
                type="button"
                onClick={markAllAsRead}
                className="text-xs text-muted-foreground hover:text-accent-brand flex items-center gap-1 cursor-pointer transition-colors"
                title={t("markAllRead")}
              >
                <CheckCheck className="size-3.5" />
                <span className="hidden sm:inline">{t("markAllRead")}</span>
              </button>
            )}
          </div>

          {/* List Content */}
          <div className="max-h-[380px] overflow-y-auto divide-y divide-border/40 scrollbar-thin">
            {loading && notifications.length === 0 ? (
              <div className="py-12 text-center text-muted-foreground text-xs">
                <div className="size-6 border-2 border-accent-brand border-t-transparent rounded-full animate-spin mx-auto mb-2" />
                {locale === "ar" ? "جارِ التحميل..." : "Loading notifications..."}
              </div>
            ) : notifications.length === 0 ? (
              <div className="py-12 px-4 text-center">
                <div className="size-12 rounded-full bg-muted/50 flex items-center justify-center mx-auto mb-3 text-muted-foreground">
                  <Bell className="size-6 stroke-[1.5]" />
                </div>
                <p className="text-sm font-medium text-foreground">
                  {t("noNotifications")}
                </p>
                <p className="text-xs text-muted-foreground mt-1">
                  {locale === "ar"
                    ? "ستظهر هنا التنبيهات الخاصة بطلباتك وعروضك"
                    : "Order updates and account alerts will appear here"}
                </p>
              </div>
            ) : (
              notifications.map((item) => (
                <div
                  key={item.id}
                  onClick={() => handleNotificationClick(item)}
                  className={`p-3.5 flex items-start gap-3 transition-colors cursor-pointer group hover:bg-secondary/50 relative ${
                    !item.isRead ? "bg-accent-brand/[0.03]" : ""
                  }`}
                >
                  {/* Category Icon */}
                  <div className="size-8 rounded-xl bg-secondary flex items-center justify-center shrink-0 mt-0.5">
                    {getCategoryIcon(item.category)}
                  </div>

                  {/* Body */}
                  <div className="flex-1 min-w-0 pr-4 rtl:pr-0 rtl:pl-4">
                    <div className="flex items-center gap-1.5">
                      <p
                        className={`text-xs truncate ${
                          !item.isRead
                            ? "font-semibold text-foreground"
                            : "font-medium text-foreground/80"
                        }`}
                      >
                        {item.title}
                      </p>
                      {!item.isRead && (
                        <span className="size-1.5 rounded-full bg-accent-brand shrink-0" />
                      )}
                    </div>
                    <p className="text-[11px] text-muted-foreground line-clamp-2 mt-0.5 leading-relaxed">
                      {item.message}
                    </p>
                    <div className="flex items-center gap-2 mt-1.5 text-[10px] text-muted-foreground/70">
                      <span>
                        {new Date(item.createdAt).toLocaleDateString(locale, {
                          month: "short",
                          day: "numeric",
                          hour: "2-digit",
                          minute: "2-digit",
                        })}
                      </span>
                      {item.actionUrl && (
                        <span className="flex items-center gap-0.5 text-accent-brand">
                          <ExternalLink className="size-2.5" />
                        </span>
                      )}
                    </div>
                  </div>

                  {/* Delete button on hover */}
                  <button
                    type="button"
                    onClick={(e) => {
                      e.stopPropagation();
                      deleteNotification(item.id);
                    }}
                    className="opacity-0 group-hover:opacity-100 p-1 text-muted-foreground/60 hover:text-rose-500 rounded-lg transition-all absolute top-3 ltr:right-3 rtl:left-3"
                    title={locale === "ar" ? "حذف التنبيه" : "Dismiss"}
                  >
                    <X className="size-3.5" />
                  </button>
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}
