"use client";

import { useState } from "react";
import { useRouter, usePathname } from "@/i18n/navigation";
import { Layers, Store, Building2, ShieldCheck, X, ChevronUp, Compass } from "lucide-react";

export function PortalSwitcher() {
  const [isOpen, setIsOpen] = useState(false);
  const router = useRouter();
  const pathname = usePathname();

  let activePortal = { name: "Storefront", color: "bg-indigo-500", icon: Store };
  if (pathname.startsWith("/vendor")) {
    activePortal = { name: "Vendor Hub", color: "bg-emerald-500", icon: Building2 };
  } else if (pathname.startsWith("/admin")) {
    activePortal = { name: "Admin Console", color: "bg-amber-500", icon: ShieldCheck };
  } else if (pathname === "/portal") {
    activePortal = { name: "Route Manager", color: "bg-accent-brand", icon: Layers };
  }

  const portals = [
    {
      name: "Public Storefront",
      path: "/",
      icon: Store,
      active: !pathname.startsWith("/vendor") && !pathname.startsWith("/admin") && pathname !== "/portal"
    },
    {
      name: "Vendor Hub",
      path: "/vendor",
      icon: Building2,
      active: pathname.startsWith("/vendor")
    },
    {
      name: "Admin Control Panel",
      path: "/admin",
      icon: ShieldCheck,
      active: pathname.startsWith("/admin")
    },
    {
      name: "Route Manager",
      path: "/portal",
      icon: Compass,
      active: pathname === "/portal"
    }
  ];

  return (
    <div className="fixed bottom-4 left-4 z-50 font-sans">
      {isOpen && (
        <div className="mb-2 w-56 rounded-2xl bg-card border border-border shadow-2xl p-2.5 animate-in fade-in slide-in-from-bottom-2 duration-200">
          <div className="flex items-center justify-between px-2 py-1.5 border-b border-border/60 mb-1.5">
            <div className="flex items-center gap-1.5 text-xs font-bold text-foreground">
              <Layers className="size-3.5 text-accent-brand" />
              <span>Route Switcher</span>
            </div>
            <button
              type="button"
              onClick={() => setIsOpen(false)}
              className="text-muted-foreground hover:text-foreground p-0.5 rounded-md hover:bg-secondary cursor-pointer"
            >
              <X className="size-3.5" />
            </button>
          </div>

          <div className="space-y-1">
            {portals.map((portal) => {
              const Icon = portal.icon;
              return (
                <button
                  type="button"
                  key={portal.path}
                  onClick={() => {
                    router.push(portal.path);
                    setIsOpen(false);
                  }}
                  className={`w-full text-left px-3 py-2 rounded-xl text-xs font-semibold flex items-center justify-between transition-colors cursor-pointer ${
                    portal.active
                      ? "bg-accent-brand text-white shadow-xs"
                      : "text-muted-foreground hover:text-foreground hover:bg-secondary"
                  }`}
                >
                  <div className="flex items-center gap-2">
                    <Icon className="size-3.5" />
                    <span>{portal.name}</span>
                  </div>
                  {portal.active && <span className="size-1.5 rounded-full bg-white" />}
                </button>
              );
            })}
          </div>
        </div>
      )}

      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 px-3 py-2 rounded-full bg-foreground text-background shadow-xl hover:opacity-90 transition-all active:scale-95 text-xs font-bold border border-background/20 cursor-pointer"
        title="Quick Route Switcher"
      >
        <span className={`size-2 rounded-full ${activePortal.color} animate-pulse`} />
        <span>{activePortal.name}</span>
        <ChevronUp className={`size-3.5 transition-transform duration-200 ${isOpen ? "rotate-180" : ""}`} />
      </button>
    </div>
  );
}
