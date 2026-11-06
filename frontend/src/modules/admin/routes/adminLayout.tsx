import { useState, useRef, useEffect, useCallback } from "react";
import { Link, useRouter, usePathname } from "@/shared/i18n/navigation";
import { useAuthContext } from "@/modules/auth/model/authContext";
import {
  LayoutDashboard,
  Package,
  FolderTree,
  Tag,
  Image,
  Boxes,
  ShoppingBag,
  Users,
  CreditCard,
  Percent,
  Settings,
  Menu,
  X,
  ChevronDown,
  LogOut,
  ShieldCheck,
  ScrollText,
  BarChart3,
  MonitorSmartphone,
  Sun,
  Moon,
  Languages,
  Store,
  type LucideIcon,
} from "lucide-react";
import { useTheme } from "@/shared/theme";
import { useLocale, useTranslations } from "next-intl";
import { rememberLocaleChoice } from "@/shared/lib/localePreference";
import { adminService } from "@/modules/admin/api/adminService";
import { notify } from "@/shared/lib/notify";
import { NotificationsDropdown } from "@/shared/ui/notificationsDropdown";
import { getBaseUrl } from "@/shared/api/client";

export function AdminLayout({ children }: { children?: React.ReactNode }) {
  const { user, logout } = useAuthContext();
  const { theme, toggle: toggleTheme } = useTheme();
  const locale = useLocale();
  const t = useTranslations("admin");
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const userMenuRef = useRef<HTMLDivElement>(null);
  const navigate = useRouter();
  const pathname = usePathname();
  const location = { pathname };

  const [pendingOrdersCount, setPendingOrdersCount] = useState<number | null>(null);

  const toggleLanguage = () => {
    const nextLocale = locale === "en" ? "ar" : "en";
    rememberLocaleChoice(nextLocale);
    navigate.replace(pathname, { locale: nextLocale });
  };

  // Play subtle audible chime on new incoming order
  const playAlertSound = useCallback(() => {
    try {
      const AudioCtx = window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext;
      if (!AudioCtx) return;
      const ctx = new AudioCtx();
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();
      osc.type = "sine";
      osc.frequency.setValueAtTime(587.33, ctx.currentTime); // D5
      osc.frequency.exponentialRampToValueAtTime(880, ctx.currentTime + 0.15); // A5
      gain.gain.setValueAtTime(0.15, ctx.currentTime);
      gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.3);
      osc.connect(gain);
      gain.connect(ctx.destination);
      osc.start();
      osc.stop(ctx.currentTime + 0.3);
    } catch {
      // Audio autoplay policy quiet bypass
    }
  }, []);

  // Initial load & refresher of dashboard aggregates (pending count)
  const refreshStats = useCallback(() => {
    adminService.getDashboard().then((res) => {
      if (res?.data) {
        setPendingOrdersCount(res.data.pendingOrders ?? 0);
      }
    }).catch(() => {});
  }, []);

  useEffect(() => {
    refreshStats();

    const handleRefresh = () => {
      refreshStats();
    };

    window.addEventListener("refresh-admin-stats", handleRefresh);
    return () => {
      window.removeEventListener("refresh-admin-stats", handleRefresh);
    };
  }, [refreshStats]);

  // Realtime Server-Sent Events (SSE) listener for instantaneous on-spot order push
  useEffect(() => {
    if (typeof window === "undefined") return;

    const streamUrl = `${getBaseUrl()}/api/v1/admin/events`;
    const eventSource = new EventSource(streamUrl, { withCredentials: true });

    eventSource.addEventListener("order.created", (e: MessageEvent) => {
      try {
        const orderData = JSON.parse(e.data);
        const orderId = orderData.id;
        const orderNum = orderData.orderNumber || `PRIM-${String(orderId).slice(0, 8).toUpperCase()}`;
        const customerName = orderData.customerName || (orderData.customerEmail ? orderData.customerEmail.split("@")[0] : "Customer");
        const formattedAmount = `EGP ${((orderData.totalAmount || 0) / 100).toFixed(2)}`;

        // Increment dynamic pending count
        setPendingOrdersCount((prev) => (prev !== null ? prev + 1 : 1));

        // Dispatch local event for orders view and notifications bell
        window.dispatchEvent(new CustomEvent("admin-order-placed", { detail: orderData }));
        window.dispatchEvent(new CustomEvent("refresh-notifications"));

        // Trigger audible ping
        playAlertSound();

        // Show immediate on-spot notification with direct link to the order
        notify.info(
          t("orders.toast.newOrderNotification", {
            orderNumber: orderNum,
            customer: customerName,
            amount: formattedAmount,
          }),
          {
            duration: 9000,
            action: {
              label: t("orders.toast.viewOrderAction"),
              onClick: () => navigate.push(`/admin/orders/${orderId}`),
            },
          }
        );
      } catch {
        // Ignore unparseable frames
      }
    });

    eventSource.addEventListener("stock.updated", (e: MessageEvent) => {
      try {
        const stockData = JSON.parse(e.data);

        // Dispatch local event for inventory table & audit log refresh
        window.dispatchEvent(new CustomEvent("admin-stock-updated", { detail: stockData }));

        // If the item reached low stock threshold (<= 10), trigger warning toast & audible alert
        if (stockData.isLowStock) {
          playAlertSound();
          const displayName = stockData.variantTitle && stockData.variantTitle !== "Default"
            ? `${stockData.productName} (${stockData.variantTitle})`
            : stockData.productName;

          notify.error(
            t("inventory.toast.lowStockAlert", {
              name: displayName,
              sku: stockData.sku || "—",
              count: stockData.currentStock,
            }),
            {
              duration: 9000,
              action: {
                label: t("inventory.toast.viewInventoryAction"),
                onClick: () => navigate.push("/admin/inventory"),
              },
            }
          );
        }
      } catch {
        // Ignore unparseable frames
      }
    });

    return () => {
      eventSource.close();
    };
  }, [navigate, playAlertSound, t]);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (userMenuRef.current && !userMenuRef.current.contains(e.target as Node)) {
        setUserMenuOpen(false);
      }
    };
    if (userMenuOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [userMenuOpen]);

  const handleLogout = async () => {
    setUserMenuOpen(false);
    await logout();
    navigate.push("/");
  };

  const navItems: {
    nameKey: string;
    path: string;
    icon: LucideIcon;
    group: "overview" | "catalog" | "sales" | "people" | "system";
    exact?: boolean;
    badgeKey?: string;
    superadminOnly?: boolean;
  }[] = [
    { nameKey: "layout.nav.dashboard", path: "/admin", icon: LayoutDashboard, group: "overview", exact: true },
    { nameKey: "layout.nav.reports", path: "/admin/reports", icon: BarChart3, group: "overview" },
    { nameKey: "layout.nav.orders", path: "/admin/orders", icon: ShoppingBag, group: "sales", badgeKey: "layout.badges.newOrders" },
    { nameKey: "layout.nav.payments", path: "/admin/payments", icon: CreditCard, group: "sales" },
    { nameKey: "layout.nav.promotions", path: "/admin/promotions", icon: Percent, group: "sales" },
    { nameKey: "layout.nav.products", path: "/admin/products", icon: Package, group: "catalog" },
    { nameKey: "layout.nav.categories", path: "/admin/categories", icon: FolderTree, group: "catalog" },
    { nameKey: "layout.nav.brands", path: "/admin/brands", icon: Tag, group: "catalog" },
    { nameKey: "layout.nav.inventory", path: "/admin/inventory", icon: Boxes, group: "catalog" },
    { nameKey: "layout.nav.mediaLibrary", path: "/admin/media", icon: Image, group: "catalog" },
    { nameKey: "layout.nav.usersStaff", path: "/admin/customers", icon: Users, group: "people" },
    { nameKey: "layout.nav.sellers", path: "/admin/sellers", icon: Store, group: "people" },
    { nameKey: "layout.nav.settings", path: "/admin/settings", icon: Settings, group: "system" },
    { nameKey: "layout.nav.system", path: "/admin/system", icon: ShieldCheck, group: "system" },
    { nameKey: "layout.nav.audit", path: "/admin/audit", icon: ScrollText, group: "system", superadminOnly: true },
    { nameKey: "layout.nav.sessions", path: "/admin/sessions", icon: MonitorSmartphone, group: "system", superadminOnly: true },
  ];

  const visibleNavItems = navItems.filter(
    (item) => !item.superadminOnly || user?.role === "superadmin"
  );

  const navGroups = [
    { key: "overview", items: visibleNavItems.filter((item) => item.group === "overview") },
    { key: "sales", items: visibleNavItems.filter((item) => item.group === "sales") },
    { key: "catalog", items: visibleNavItems.filter((item) => item.group === "catalog") },
    { key: "people", items: visibleNavItems.filter((item) => item.group === "people") },
    { key: "system", items: visibleNavItems.filter((item) => item.group === "system") },
  ].filter((group) => group.items.length > 0);

  return (
    <div className="min-h-screen bg-muted/20 text-foreground flex flex-col font-sans">
      {/* Top Bar */}
      <header className="sticky top-0 z-40 w-full bg-background border-b border-border h-14 px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="p-2 rounded-lg hover:bg-secondary lg:hidden text-muted-foreground hover:text-foreground"
            aria-label={t("layout.toggleNav")}
          >
            {sidebarOpen ? <X className="size-5" /> : <Menu className="size-5" />}
          </button>

          <Link href="/" className="flex items-center gap-2 group hover:opacity-90 transition-opacity">
            <span dir="ltr" className="font-black text-lg tracking-tighter text-foreground">
              PRI<span className="text-accent-brand">M</span>
            </span>
            <span className="hidden border-s border-border ps-3 text-xs font-medium text-muted-foreground sm:inline">
              {t("layout.brandTitle")}
            </span>
          </Link>
        </div>

        <div className="flex items-center gap-3">
          <NotificationsDropdown />

          {/* Account Dropdown */}
          <div className="relative" ref={userMenuRef}>
            <button
              type="button"
              onClick={() => setUserMenuOpen((prev) => !prev)}
              className="flex items-center gap-2.5 px-2 py-1.5 rounded-xl hover:bg-secondary/70 transition-colors cursor-pointer"
            >
              <div className="size-8 rounded-full bg-accent-brand/10 border border-accent-brand/30 text-accent-brand flex items-center justify-center font-bold text-xs uppercase">
                {user?.email ? user.email.slice(0, 2) : "SA"}
              </div>
              <div className="hidden sm:flex flex-col text-start">
                <span className="text-xs font-semibold text-foreground leading-none">
                  {user?.name || (user?.role === "superadmin" ? t("roles.superadmin") : t("roles.administrator"))}
                </span>
                <span className="text-[10px] text-muted-foreground">{user?.email || "superadmin@prim.com"}</span>
              </div>
              <ChevronDown className="size-3.5 text-muted-foreground ms-1" />
            </button>

            {userMenuOpen && (
              <div className="absolute right-0 rtl:left-0 rtl:right-auto mt-2 w-64 bg-background border border-border rounded-2xl shadow-xl z-50 overflow-hidden py-1.5 animate-in fade-in-0 zoom-in-95 duration-100">
                <div className="px-4 py-3 border-b border-border/70 bg-secondary/20">
                  <div className="flex items-center justify-between">
                    <p className="font-semibold text-xs text-foreground truncate">{user?.name || t("roles.superadmin")}</p>
                    <span className="text-[10px] text-muted-foreground uppercase tracking-wider">
                      {user?.role || "superadmin"}
                    </span>
                  </div>
                  <p className="text-[11px] text-muted-foreground truncate mt-0.5">{user?.email || "superadmin@prim.com"}</p>
                </div>

                <div className="space-y-0.5 p-1.5 text-sm">
                  <button
                    type="button"
                    onClick={toggleTheme}
                    className="flex w-full cursor-pointer items-center justify-between rounded-md px-3 py-2 text-start text-foreground transition-colors hover:bg-secondary/70"
                  >
                    <span className="flex items-center gap-2.5">
                      {theme === "light" ? <Moon className="size-4 text-muted-foreground" /> : <Sun className="size-4 text-muted-foreground" />}
                      <span>{t("layout.theme")}</span>
                    </span>
                    <span className="text-xs capitalize text-muted-foreground">{theme}</span>
                  </button>

                  <button
                    type="button"
                    onClick={toggleLanguage}
                    className="flex w-full cursor-pointer items-center justify-between rounded-md px-3 py-2 text-start text-foreground transition-colors hover:bg-secondary/70"
                  >
                    <span className="flex items-center gap-2.5">
                      <Languages className="size-4 text-muted-foreground" />
                      <span>{t("layout.language")}</span>
                    </span>
                    <span className="text-xs text-muted-foreground">{locale === "en" ? "العربية" : "English"}</span>
                  </button>

                  <Link
                    href="/admin/settings"
                    onClick={() => setUserMenuOpen(false)}
                    className="flex items-center gap-2.5 rounded-md px-3 py-2 text-foreground transition-colors hover:bg-secondary/70"
                  >
                    <Settings className="size-4 shrink-0 text-muted-foreground" />
                    <span>{t("layout.storeConfiguration")}</span>
                  </Link>
                </div>

                <div className="border-t border-border p-1.5">
                  <button
                    onClick={handleLogout}
                    className="flex w-full cursor-pointer items-center gap-2.5 rounded-md px-3 py-2 text-sm font-medium text-destructive transition-colors hover:bg-destructive/10"
                  >
                    <LogOut className="size-4 shrink-0" />
                    <span>{t("layout.signOut")}</span>
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </header>

      <div className="flex flex-1 relative">
        {/* Mobile Backdrop */}
        {sidebarOpen && (
          <div
            onClick={() => setSidebarOpen(false)}
            className="fixed inset-0 z-30 bg-black/40 backdrop-blur-xs lg:hidden"
          ></div>
        )}

        {/* Sidebar */}
        <aside
          className={`fixed start-0 top-14 z-30 flex h-[calc(100dvh-3.5rem)] w-60 flex-col border-e border-border bg-sidebar transition-transform duration-200 ease-out lg:sticky ${
            sidebarOpen ? "translate-x-0" : "max-lg:-translate-x-full max-lg:rtl:translate-x-full"
          }`}
        >
          <nav className="flex-1 space-y-5 overflow-y-auto p-3">
            {navGroups.map((group) => (
              <div key={group.key} className="space-y-0.5">
                <div className="px-3 pb-1 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
                  {t(`layout.groups.${group.key}`)}
                </div>
                {group.items.map((item) => {
                  const Icon = item.icon;
                  const isActive = item.exact
                    ? location.pathname === item.path
                    : location.pathname.startsWith(item.path);
                  const showBadge =
                    item.badgeKey === "layout.badges.newOrders" &&
                    pendingOrdersCount !== null &&
                    pendingOrdersCount > 0;
                  return (
                    <Link
                      key={item.path}
                      href={item.path}
                      onClick={() => setSidebarOpen(false)}
                      aria-current={isActive ? "page" : undefined}
                      className={`flex items-center justify-between gap-3 rounded-md px-3 py-2 text-sm transition-colors ${
                        isActive
                          ? "bg-background font-medium text-foreground shadow-xs"
                          : "text-muted-foreground hover:bg-background/70 hover:text-foreground"
                      }`}
                    >
                      <span className="flex min-w-0 items-center gap-2.5">
                        <Icon className="size-4 shrink-0" />
                        <span className="truncate">{t(item.nameKey)}</span>
                      </span>
                      {showBadge && item.badgeKey ? (
                        <span className="rounded-md border border-border bg-secondary/70 px-1.5 py-0.5 text-[11px] font-medium tabular-nums text-muted-foreground">
                          {t(item.badgeKey, { count: pendingOrdersCount })}
                        </span>
                      ) : null}
                    </Link>
                  );
                })}
              </div>
            ))}
          </nav>
        </aside>

        {/* Main Workspace Area */}
        <main className="min-w-0 flex-1 overflow-y-auto p-4 md:p-6">
          <div className="mx-auto w-full max-w-[1400px]">{children}</div>
        </main>
      </div>
    </div>
  );
}
