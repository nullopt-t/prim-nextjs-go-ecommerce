"use client";

import { useRouter } from "@/i18n/navigation";
import { Store, Building2, ShieldCheck, ArrowRight, Sparkles, Layers, CheckCircle2 } from "lucide-react";

export function PortalManager() {
  const navigate = useRouter();

  const portals = [
    {
      id: "public",
      title: "Public Storefront",
      subtitle: "Customer Shopping Experience",
      description: "Explore the live public marketplace, browse products, add to cart, and test the checkout workflow.",
      icon: Store,
      badge: "Customer Facing",
      color: "from-blue-500/10 to-indigo-500/10 text-indigo-600 dark:text-indigo-400 border-indigo-500/20 hover:border-indigo-500/50",
      iconBg: "bg-indigo-500 text-white",
      buttonColor: "bg-indigo-600 hover:bg-indigo-700 text-white",
      route: "/home",
      features: [
        "Product catalog with filter & search",
        "Cart drawer & real-time pricing",
        "Multilingual (EN / AR) support",
        "Wishlist & user account dashboard"
      ]
    },
    {
      id: "vendor",
      title: "Vendor Hub",
      subtitle: "Merchant & Partner Portal",
      description: "Dedicated portal for independent seller shops to list items, monitor stock levels, view orders, and manage payouts.",
      icon: Building2,
      badge: "Merchant Portal",
      color: "from-emerald-500/10 to-teal-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20 hover:border-emerald-500/50",
      iconBg: "bg-emerald-500 text-white",
      buttonColor: "bg-emerald-600 hover:bg-emerald-700 text-white",
      route: "/vendor",
      features: [
        "Vendor analytics & sales revenue",
        "Product submission & SKU management",
        "Stock inventory & low-stock alerts",
        "Payout requests & bank details"
      ]
    },
    {
      id: "admin",
      title: "Admin Control Panel",
      subtitle: "Enterprise Platform Operations",
      description: "Full store management console for platform administrators to manage products, categories, orders, customers, and site settings.",
      icon: ShieldCheck,
      badge: "System Operations",
      color: "from-amber-500/10 to-rose-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20 hover:border-amber-500/50",
      iconBg: "bg-accent-brand text-white",
      buttonColor: "bg-accent-brand hover:bg-accent-brand/90 text-white",
      route: "/admin",
      features: [
        "System-wide sales performance overview",
        "Catalog category & attribute builder",
        "Customer management & order fulfillment",
        "Payment gateway & promotion rules"
      ]
    }
  ];

  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col font-sans">
      {/* Top Header */}
      <header className="border-b border-border bg-background/80 backdrop-blur-md sticky top-0 z-30 h-16 px-6 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="size-9 rounded-xl bg-accent-brand flex items-center justify-center text-white font-bold text-lg shadow-sm">
            P
          </div>
          <div>
            <h1 className="font-bold text-base leading-tight">PRIM Route Manager</h1>
            <p className="text-xs text-muted-foreground">Business Portal Switcher</p>
          </div>
        </div>

        <div className="flex items-center gap-2 text-xs font-semibold px-3 py-1.5 rounded-full bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
          <CheckCircle2 className="size-3.5" />
          <span>All 3 Environments Active</span>
        </div>
      </header>

      {/* Hero Section */}
      <main className="flex-1 max-w-6xl w-full mx-auto px-4 sm:px-6 py-12 flex flex-col justify-center">
        <div className="text-center max-w-2xl mx-auto mb-12 space-y-3">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-accent-brand/10 text-accent-brand text-xs font-semibold border border-accent-brand/20 mb-1">
            <Layers className="size-3.5" />
            <span>Multi-Portal Architecture</span>
          </div>
          <h2 className="text-3xl sm:text-4xl font-extrabold tracking-tight">
            Select Your Destination
          </h2>
          <p className="text-muted-foreground text-sm sm:text-base leading-relaxed">
            The public customer storefront is segregated from internal operations. Choose which workspace environment you wish to launch:
          </p>
        </div>

        {/* Portal Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 lg:gap-8">
          {portals.map((portal) => {
            const Icon = portal.icon;
            return (
              <div
                key={portal.id}
                className="group relative rounded-2xl border border-border bg-card p-6 shadow-xs hover:shadow-xl transition-all duration-300 flex flex-col justify-between hover:-translate-y-1"
              >
                <div>
                  <div className="flex items-center justify-between mb-5">
                    <div className={`size-12 rounded-xl ${portal.iconBg} flex items-center justify-center shadow-md group-hover:scale-105 transition-transform`}>
                      <Icon className="size-6" />
                    </div>
                    <span className="text-[11px] font-bold px-2.5 py-1 rounded-full bg-secondary text-secondary-foreground border border-border">
                      {portal.badge}
                    </span>
                  </div>

                  <h3 className="text-xl font-bold text-foreground mb-1 group-hover:text-accent-brand transition-colors">
                    {portal.title}
                  </h3>
                  <p className="text-xs font-semibold text-accent-brand mb-3">
                    {portal.subtitle}
                  </p>
                  <p className="text-xs text-muted-foreground leading-relaxed mb-6">
                    {portal.description}
                  </p>

                  <div className="space-y-2 border-t border-border/60 pt-4 mb-6">
                    <p className="text-[11px] font-bold text-foreground uppercase tracking-wider">Included Features:</p>
                    {portal.features.map((feat, i) => (
                      <div key={i} className="flex items-center gap-2 text-xs text-muted-foreground">
                        <span className="size-1.5 rounded-full bg-accent-brand shrink-0"></span>
                        <span>{feat}</span>
                      </div>
                    ))}
                  </div>
                </div>

                <button
                  onClick={() => navigate.push(portal.route)}
                  className={`w-full py-3 px-4 rounded-xl text-xs font-bold ${portal.buttonColor} flex items-center justify-center gap-2 shadow-sm transition-all duration-200 cursor-pointer`}
                >
                  <span>Launch {portal.title}</span>
                  <ArrowRight className="size-4 group-hover:translate-x-1 transition-transform" />
                </button>
              </div>
            );
          })}
        </div>

        {/* Footer Note */}
        <div className="mt-12 p-4 rounded-xl border border-border bg-secondary/30 text-center text-xs text-muted-foreground max-w-xl mx-auto flex items-center justify-center gap-2">
          <Sparkles className="size-4 text-accent-brand shrink-0" />
          <span>You can return to this Route Manager anytime at <code className="px-1.5 py-0.5 rounded bg-muted text-foreground font-mono">/portal</code> or via the floating Portal Switcher.</span>
        </div>
      </main>
    </div>
  );
}
