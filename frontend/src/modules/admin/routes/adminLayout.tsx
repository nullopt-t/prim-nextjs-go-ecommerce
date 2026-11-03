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
  User,
  LogOut,
  ShieldCheck,
  ScrollText,
  BarChart3,
  MonitorSmartphone,
  Sun,
  Moon,
  Languages,
  Bell,
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
    exact?: boolean;
    badge?: string;
    badgeKey?: string;
    superadminOnly?: boolean;
  }[] = [
    { nameKey: "layout.nav.dashboard", path: "/admin", icon: LayoutDashboard, exact: true },
    {
      nameKey: "layout.nav.orders",
      path: "/admin/orders",
      icon: ShoppingBag,
      badge: pendingOrdersCount !== null && pendingOrdersCount > 0 ? `${pendingOrdersCount} New` : undefined,
      badgeKey: pendingOrdersCount !== null && pendingOrdersCount > 0 ? "layout.badges.newOrders" : undefined,
    },
    { nameKey: "layout.nav.products", path: "/admin/products", icon: Package },
    { nameKey: "layout.nav.categories", path: "/admin/categories", icon: FolderTree },
    { nameKey: "layout.nav.brands", path: "/admin/brands", icon: Tag },
    { nameKey: "layout.nav.inventory", path: "/admin/inventory", icon: Boxes, badge: "Low Stock", badgeKey: "layout.badges.lowStock" },
    { nameKey: "layout.nav.mediaLibrary", path: "/admin/media", icon: Image },
    { nameKey: "layout.nav.usersStaff", path: "/admin/customers", icon: Users },
    { nameKey: "layout.nav.payments", path: "/admin/payments", icon: CreditCard },
    { nameKey: "layout.nav.promotions", path: "/admin/promotions", icon: Percent },
    { nameKey: "layout.nav.reports", path: "/admin/reports", icon: BarChart3 },
    { nameKey: "layout.nav.settings", path: "/admin/settings", icon: Settings },
    { nameKey: "layout.nav.system", path: "/admin/system", icon: ShieldCheck },
    { nameKey: "layout.nav.audit", path: "/admin/audit", icon: ScrollText, superadminOnly: true },
    { nameKey: "layout.nav.sessions", path: "/admin/sessions", icon: MonitorSmartphone, superadminOnly: true },
  ];

  const visibleNavItems = navItems.filter(
    (item) => !item.superadminOnly || user?.role === "superadmin"
  );

  return (
    <div className="min-h-screen bg-muted/20 text-foreground flex flex-col font-sans">
      {/* Top Bar */}
      <header className="sticky top-0 z-40 w-full bg-background border-b border-border h-16 px-4 md:px-6 flex items-center justify-between shadow-xs">
        <div className="flex items-center gap-3">
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="p-2 rounded-lg hover:bg-secondary lg:hidden text-muted-foreground hover:text-foreground"
            aria-label={t("layout.toggleNav")}
          >
            {sidebarOpen ? <X className="size-5" /> : <Menu className="size-5" />}
          </button>

          <Link href="/" className="flex items-center gap-2 group hover:opacity-90 transition-opacity">
            <span dir="ltr" className="font-black text-xl tracking-tighter text-foreground">
              PRI<span className="text-accent-brand">M</span>
            </span>
            <div className="hidden sm:flex flex-col">
              <span className="font-bold text-xs leading-tight text-foreground">{t("layout.brandTitle")}</span>
              <span className="text-[9px] text-muted-foreground font-medium uppercase tracking-wider">{t("layout.brandSubtitle")}</span>
            </div>
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

                <div className="p-1.5 space-y-0.5 text-xs font-medium">
                  <button
                    type="button"
                    onClick={() => {
                      toggleTheme();
                    }}
                    className="w-full flex items-center justify-between px-3 py-2 rounded-xl text-foreground hover:text-accent-brand hover:bg-secondary/70 transition-colors cursor-pointer text-start"
                  >
                    <div className="flex items-center gap-2.5">
                      {theme === "light" ? <Moon className="size-4 text-muted-foreground" /> : <Sun className="size-4 text-muted-foreground" />}
                      <span>{t("layout.theme")}</span>
                    </div>
                    <span className="text-[11px] font-semibold text-muted-foreground capitalize">{theme}</span>
                  </button>

                  <button
                    type="button"
                    onClick={() => {
                      toggleLanguage();
                    }}
                    className="w-full flex items-center justify-between px-3 py-2 rounded-xl text-foreground hover:text-accent-brand hover:bg-secondary/70 transition-colors cursor-pointer text-start"
                  >
                    <div className="flex items-center gap-2.5">
                      <Languages className="size-4 text-muted-foreground" />
                      <span>{t("layout.language")}</span>
                    </div>
                    <span className="text-[11px] font-semibold text-muted-foreground uppercase">{locale === "en" ? "العربية" : "English"}</span>
                  </button>

                  <div className="my-1 border-t border-border/50"></div>

                  <Link
                    href="/user/overview"
                    onClick={() => setUserMenuOpen(false)}
                    className="flex items-center gap-2.5 px-3 py-2 rounded-xl text-foreground hover:text-accent-brand hover:bg-secondary/70 transition-colors"
                  >
                    <User className="size-4 text-muted-foreground shrink-0" />
                    <span>{t("layout.customerAccountHub")}</span>
                  </Link>

                  <Link
                    href="/admin/settings"
                    onClick={() => setUserMenuOpen(false)}
                    className="flex items-center gap-2.5 px-3 py-2 rounded-xl text-foreground hover:text-accent-brand hover:bg-secondary/70 transition-colors"
                  >
                    <Settings className="size-4 text-muted-foreground shrink-0" />
                    <span>{t("layout.storeConfiguration")}</span>
                  </Link>

                  <Link
                    href="/admin/customers"
                    onClick={() => setUserMenuOpen(false)}
                    className="flex items-center gap-2.5 px-3 py-2 rounded-xl text-foreground hover:text-accent-brand hover:bg-secondary/70 transition-colors"
                  >
                    <Users className="size-4 text-muted-foreground shrink-0" />
                    <span>{t("layout.userManagement")}</span>
                  </Link>
                </div>

                <div className="p-1.5 border-t border-border/70">
                  <button
                    onClick={handleLogout}
                    className="w-full flex items-center gap-2.5 px-3 py-2 rounded-xl text-destructive hover:bg-destructive/10 transition-colors text-xs font-semibold cursor-pointer"
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
          className={`fixed lg:sticky start-0 top-16 z-30 h-[calc(100dvh-4rem)] w-64 bg-background border-e border-border flex flex-col transition-transform duration-200 ease-in-out ${
            sidebarOpen ? "translate-x-0" : "max-lg:-translate-x-full max-lg:rtl:translate-x-full"
          }`}
        >
          <div className="p-3 space-y-1 overflow-y-auto flex-1">
            <div className="px-3 py-2 text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
              {t("layout.management")}
            </div>

            {visibleNavItems.map((item) => {
              const Icon = item.icon;
              const isActive = item.exact ? location.pathname === item.path : location.pathname.startsWith(item.path);
              return (
                <Link
                  key={item.path}
                  href={item.path}
                  onClick={() => setSidebarOpen(false)}
                  className={`flex items-center justify-between px-3 py-2.5 rounded-xl text-xs font-medium transition-all ${
                    isActive
                      ? "bg-accent-brand text-white font-semibold shadow-xs"
                      : "text-muted-foreground hover:bg-secondary hover:text-foreground"
                  }`}
                >
                  <div className="flex items-center gap-2.5">
                    <Icon className="size-4 shrink-0" />
                    <span>{t(item.nameKey)}</span>
                  </div>
                  {item.badge && (
                    <span
                      className={`px-1.5 py-0.5 rounded-md text-[10px] font-bold ${
                        item.badge === "Low Stock"
                          ? "bg-amber-100 text-amber-800 dark:bg-amber-900/50 dark:text-amber-300"
                          : "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/50 dark:text-emerald-300"
                      }`}
                    >
                      {item.badgeKey
                        ? item.badgeKey === "layout.badges.newOrders" && pendingOrdersCount !== null
                          ? t(item.badgeKey, { count: pendingOrdersCount })
                          : t(item.badgeKey)
                        : item.badge}
                    </span>
                  )}
                </Link>
              );
            })}
          </div>
        </aside>

        {/* Main Workspace Area */}
        <main className="flex-1 p-4 md:p-6 lg:p-8 min-w-0 overflow-y-auto">
          {children}
        </main>
      </div>
    </div>
  );
}
