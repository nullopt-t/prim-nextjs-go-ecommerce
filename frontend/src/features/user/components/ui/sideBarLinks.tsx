"use client";

import { usePathname, useRouter, Link } from "@/i18n/navigation";
import { useAuthContext } from "@/context/AuthContext";
import {
  Blocks,
  Package,
  Heart,
  MapPin,
  CreditCard,
  Star,
  Settings,
  LogOut,
} from "lucide-react";

export const navLinks = [
  { id: "overview", icon: Blocks, label: "Overview", path: "/user/overview" },
  { id: "orders", icon: Package, label: "My Orders", path: "/user/orders" },
  { id: "wishlist", icon: Heart, label: "Wishlist", path: "/user/wishlist" },
  { id: "address", icon: MapPin, label: "Addresses", path: "/user/address" },
  { id: "payment", icon: CreditCard, label: "Payment", path: "/user/payment" },
  { id: "reviews", icon: Star, label: "Reviews", path: "/user/reviews" },
  { id: "settings", icon: Settings, label: "Settings", path: "/user/settings" },
];

export default function UserSideBarLinks() {
  const pathname = usePathname();
  const router = useRouter();
  const { logout } = useAuthContext();

  const handleLogout = async () => {
    await logout();
    router.push("/auth");
  };

  return (
    <ul className="flex flex-col gap-0.5">
      {navLinks.map((item) => {
        const isActive = pathname === item.path;
        return (
          <li key={item.id}>
            <Link
              href={item.path}
              className={`flex items-center gap-3 px-3 py-2.5 rounded-md text-sm transition-colors ${
                isActive
                  ? "bg-accent-brand/10 text-accent-brand font-medium"
                  : "text-sidebar-foreground hover:bg-sidebar-accent hover:text-foreground"
              }`}
            >
              <item.icon className="size-4 shrink-0" />
              <span>{item.label}</span>
            </Link>
          </li>
        );
      })}

      {/* Logout */}
      <li className="border-t border-border mt-2 pt-2">
        <button
          type="button"
          onClick={handleLogout}
          className="w-full flex items-center gap-3 px-3 py-2.5 rounded-md text-sm transition-colors text-destructive hover:bg-destructive/10 cursor-pointer text-left"
        >
          <LogOut className="size-4 shrink-0" />
          <span>Logout</span>
        </button>
      </li>
    </ul>
  );
}
