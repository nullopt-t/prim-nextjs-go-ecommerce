import { useState, useRef, useEffect } from "react";
import { Link, useRouter, usePathname } from "@/i18n/navigation";
import { useAuthContext } from "@/context/AuthContext";
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
  Sun,
  Moon,
  Languages
} from "lucide-react";
import { useTheme } from "@/context/theme";
import { useLocale, useTranslations } from "next-intl";

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

  const toggleLanguage = () => {
    const nextLocale = locale === "en" ? "ar" : "en";
    navigate.replace(pathname, { locale: nextLocale });
  };

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

  const navItems = [
    { nameKey: "layout.nav.dashboard", path: "/admin", icon: LayoutDashboard, exact: true },
    { nameKey: "layout.nav.products", path: "/admin/products", icon: Package },
    { nameKey: "layout.nav.categories", path: "/admin/categories", icon: FolderTree },
    { nameKey: "layout.nav.brands", path: "/admin/brands", icon: Tag },
    { nameKey: "layout.nav.mediaLibrary", path: "/admin/media", icon: Image },
    { nameKey: "layout.nav.inventory", path: "/admin/inventory", icon: Boxes, badge: "Low Stock", badgeKey: "layout.badges.lowStock" },
    { nameKey: "layout.nav.orders", path: "/admin/orders", icon: ShoppingBag, badge: "4 New", badgeKey: "layout.badges.newOrders" },
    { nameKey: "layout.nav.usersStaff", path: "/admin/customers", icon: Users },
    { nameKey: "layout.nav.payments", path: "/admin/payments", icon: CreditCard },
    { nameKey: "layout.nav.promotions", path: "/admin/promotions", icon: Percent },
    { nameKey: "layout.nav.settings", path: "/admin/settings", icon: Settings },
  ];

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
              <div className="hidden sm:flex flex-col text-left">
                <span className="text-xs font-semibold text-foreground leading-none">
                  {user?.name || (user?.role === "superadmin" ? t("roles.superadmin") : t("roles.administrator"))}
                </span>
                <span className="text-[10px] text-muted-foreground">{user?.email || "superadmin@prim.com"}</span>
              </div>
              <ChevronDown className="size-3.5 text-muted-foreground ml-1" />
            </button>

            {userMenuOpen && (
              <div className="absolute right-0 rtl:left-0 rtl:right-auto mt-2 w-64 bg-background border border-border rounded-2xl shadow-xl z-50 overflow-hidden py-1.5 animate-in fade-in-0 zoom-in-95 duration-100">
                <div className="px-4 py-3 border-b border-border/70 bg-secondary/20">
                  <div className="flex items-center justify-between">
                    <p className="font-semibold text-xs text-foreground truncate">{user?.name || t("roles.superadmin")}</p>
                    <span className="px-1.5 py-0.5 rounded-md text-[9px] font-bold uppercase tracking-wider bg-purple-500/15 text-purple-700 dark:text-purple-300 border border-purple-500/20">
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
                    className="w-full flex items-center justify-between px-3 py-2 rounded-xl text-foreground hover:text-accent-brand hover:bg-secondary/70 transition-colors cursor-pointer text-left"
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
                    className="w-full flex items-center justify-between px-3 py-2 rounded-xl text-foreground hover:text-accent-brand hover:bg-secondary/70 transition-colors cursor-pointer text-left"
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
          className={`fixed lg:sticky top-16 z-30 h-[calc(100dvh-4rem)] w-64 bg-background border-r border-border flex flex-col justify-between transition-transform duration-200 ease-in-out ${
            sidebarOpen ? "translate-x-0" : "-translate-x-full lg:translate-x-0"
          }`}
        >
          <div className="p-3 space-y-1 overflow-y-auto max-h-[calc(100dvh-8rem)]">
            <div className="px-3 py-2 text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
              {t("layout.management")}
            </div>

            {navItems.map((item) => {
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
                      {item.badgeKey && t(item.badgeKey)}
                    </span>
                  )}
                </Link>
              );
            })}
          </div>

          <div className="p-3 border-t border-border bg-secondary/30">
            <div className="flex items-center justify-between px-3 py-2 text-xs text-muted-foreground">
              <div className="flex items-center gap-2">
                <ShieldCheck className="size-4 text-emerald-500" />
                <span>{t("layout.systemHealth")}</span>
              </div>
            </div>
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
