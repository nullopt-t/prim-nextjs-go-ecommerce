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
    <div className="md:hidden mb-4 -mx-4 px-4 overflow-x-auto">
      <div className="flex gap-1 border-b border-border pb-0 min-w-max">
        {navLinks.map((item) => {
          const isActive = pathname === item.path;
          return (
            <Link
              key={item.id}
              href={item.path}
              className={`px-4 py-2 flex-shrink-0 flex flex-col items-center gap-0.5 text-[10px] font-medium transition-colors border-b-2 -mb-px ${
                isActive
                  ? "text-accent-brand border-accent-brand"
                  : "text-muted-foreground border-transparent hover:text-foreground"
              }`}
            >
              <item.icon className="size-4" />
              <span>{item.label}</span>
            </Link>
          );
        })}
        <button
          type="button"
          onClick={handleLogout}
          className="px-4 py-2 flex-shrink-0 flex flex-col items-center gap-0.5 text-[10px] font-medium text-destructive border-b-2 border-transparent -mb-px cursor-pointer transition-colors hover:opacity-80"
        >
          <LogOut className="size-4" />
          <span>Logout</span>
        </button>
      </div>
    </div>
  );
}
