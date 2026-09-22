import { useState } from "react";
import { Link, useRouter, usePathname } from "@/i18n/navigation";
import {
  LayoutDashboard,
  Package,
  FolderTree,
  Tag,
  SlidersHorizontal,
  Image,
  Boxes,
  ShoppingBag,
  Users,
  CreditCard,
  Percent,
  Settings,
  Store,
  Menu,
  X,
  Bell,
  ChevronRight,
  ShieldCheck
} from "lucide-react";

export function AdminLayout({ children }: { children?: React.ReactNode }) {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const navigate = useRouter();
  const pathname = usePathname();
  const location = { pathname };

  const navItems = [
    { name: "Dashboard", path: "/admin", icon: LayoutDashboard, exact: true },
    { name: "Products", path: "/admin/products", icon: Package },
    { name: "Categories", path: "/admin/categories", icon: FolderTree },
    { name: "Brands", path: "/admin/brands", icon: Tag },
    { name: "Attributes", path: "/admin/attributes", icon: SlidersHorizontal },
    { name: "Media Library", path: "/admin/media", icon: Image },
    { name: "Inventory", path: "/admin/inventory", icon: Boxes, badge: "Low Stock" },
    { name: "Orders", path: "/admin/orders", icon: ShoppingBag, badge: "4 New" },
    { name: "Customers", path: "/admin/customers", icon: Users },
    { name: "Payments", path: "/admin/payments", icon: CreditCard },
    { name: "Promotions", path: "/admin/promotions", icon: Percent },
    { name: "Settings", path: "/admin/settings", icon: Settings },
  ];

  // Helper to get active page title
  const getActiveTitle = () => {
    const currentPath = location.pathname;
    if (currentPath.startsWith("/admin/orders/")) return "Order Details";
    const found = navItems.find((item) =>
      item.exact ? currentPath === item.path : currentPath.startsWith(item.path)
    );
    return found ? found.name : "Admin Control Panel";
  };

  return (
    <div className="min-h-screen bg-muted/20 text-foreground flex flex-col font-sans">
      {/* Top Bar */}
      <header className="sticky top-0 z-40 w-full bg-background border-b border-border h-16 px-4 md:px-6 flex items-center justify-between shadow-xs">
        <div className="flex items-center gap-3">
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="p-2 rounded-lg hover:bg-secondary lg:hidden text-muted-foreground hover:text-foreground"
            aria-label="Toggle Navigation"
          >
            {sidebarOpen ? <X className="size-5" /> : <Menu className="size-5" />}
          </button>

          <div className="flex items-center gap-2">
            <div className="size-8 rounded-lg bg-accent-brand flex items-center justify-center text-white font-bold text-sm shadow-xs">
              P
            </div>
            <div className="hidden sm:flex flex-col">
              <span className="font-bold text-sm leading-tight text-foreground">PRIM Admin</span>
              <span className="text-[10px] text-muted-foreground font-medium uppercase tracking-wider">Enterprise Console</span>
            </div>
          </div>

          <div className="hidden md:flex items-center gap-1 text-xs text-muted-foreground ml-6 bg-secondary/50 px-3 py-1.5 rounded-full border border-border">
            <span>Admin</span>
            <ChevronRight className="size-3" />
            <span className="font-semibold text-foreground">{getActiveTitle()}</span>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={() => navigate.push("/portal")}
            className="hidden sm:flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-lg bg-accent-brand/10 text-accent-brand hover:bg-accent-brand/20 border border-accent-brand/20 transition-colors"
          >
            <ShieldCheck className="size-3.5" />
            <span>Route Manager</span>
          </button>

          <button
            onClick={() => navigate.push("/home")}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-lg bg-secondary text-secondary-foreground hover:bg-secondary/80 border border-border transition-colors"
          >
            <Store className="size-3.5 text-accent-brand" />
            <span className="hidden sm:inline">Storefront</span>
          </button>

          <div className="h-4 w-px bg-border"></div>

          <button className="p-2 rounded-lg hover:bg-secondary text-muted-foreground hover:text-foreground relative">
            <Bell className="size-4.5" />
            <span className="absolute top-1.5 right-1.5 size-2 bg-accent-brand rounded-full"></span>
          </button>

          <div className="flex items-center gap-2.5 pl-2 border-l border-border">
            <div className="size-8 rounded-full bg-accent-brand/10 border border-accent-brand/30 text-accent-brand flex items-center justify-center font-bold text-xs">
              AD
            </div>
            <div className="hidden xl:flex flex-col text-left">
              <span className="text-xs font-semibold text-foreground leading-none">Admin Store Owner</span>
              <span className="text-[10px] text-muted-foreground">admin@aurashop.com</span>
            </div>
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
          className={`fixed lg:sticky top-16 z-30 h-[calc(100vh-4rem)] w-64 bg-background border-r border-border flex flex-col justify-between transition-transform duration-200 ease-in-out ${
            sidebarOpen ? "translate-x-0" : "-translate-x-full lg:translate-x-0"
          }`}
        >
          <div className="p-3 space-y-1 overflow-y-auto max-h-[calc(100vh-8rem)]">
            <div className="px-3 py-2 text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
              Management
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
                    <span>{item.name}</span>
                  </div>
                  {item.badge && (
                    <span
                      className={`px-1.5 py-0.5 rounded-md text-[10px] font-bold ${
                        item.badge === "Low Stock"
                          ? "bg-amber-100 text-amber-800 dark:bg-amber-900/50 dark:text-amber-300"
                          : "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/50 dark:text-emerald-300"
                      }`}
                    >
                      {item.badge}
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
                <span>System Health: 100%</span>
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
