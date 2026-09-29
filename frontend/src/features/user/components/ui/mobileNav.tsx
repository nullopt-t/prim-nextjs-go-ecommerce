"use client";

import { usePathname, useRouter, Link } from "@/i18n/navigation";
import { useAuthContext } from "@/context/AuthContext";
import { navLinks } from "@/features/user/components/ui/sideBarLinks";
import { LogOut } from "lucide-react";

export default function MobileNav() {
  const pathname = usePathname();
  const router = useRouter();
  const { logout } = useAuthContext();

  const handleLogout = async () => {
    await logout();
    router.push("/auth");
  };

  return (
    <nav className="md:hidden fixed bottom-0 inset-x-0 z-50 bg-card/95 backdrop-blur-sm border-t border-border pb-safe">
      <div className="flex items-stretch overflow-x-auto scrollbar-none">
        {navLinks.map((item) => {
          const isActive = pathname === item.path;
          return (
            <Link
              key={item.id}
              href={item.path}
              className={`flex-1 min-w-[56px] flex flex-col items-center justify-center gap-0.5 py-2.5 px-1 text-[10px] font-medium transition-colors ${
                isActive
                  ? "text-accent-brand"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              <item.icon className={`size-5 ${isActive ? "text-accent-brand" : ""}`} />
              <span className="truncate w-full text-center leading-tight">{item.label}</span>
            </Link>
          );
        })}
        <button
          type="button"
          onClick={handleLogout}
          className="flex-1 min-w-[56px] flex flex-col items-center justify-center gap-0.5 py-2.5 px-1 text-[10px] font-medium text-destructive cursor-pointer transition-colors hover:opacity-80"
        >
          <LogOut className="size-5" />
          <span className="truncate w-full text-center leading-tight">Logout</span>
        </button>
      </div>
    </nav>
  );
}
