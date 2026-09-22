"use client";

import { usePathname, Link } from "@/i18n/navigation";
import {
  Blocks,
  Box,
  Heart,
  MapPin,
  CreditCard,
  Star,
  Settings,
  LogOut,
} from "lucide-react";

export default function UserSideBarLinks() {
  const pathname = usePathname();

  const links = [
    {
      id: "overview",
      icon: Blocks,
      link: "Overview",
      path: "/user/overview",
    },
    {
      id: "orders",
      icon: Box,
      link: "My orders",
      path: "/user/orders",
    },
    {
      id: "wishlist",
      icon: Heart,
      link: "Wishlist",
      path: "/user/wishlist",
    },
    {
      id: "address",
      icon: MapPin,
      link: "Addresses",
      path: "/user/address",
    },
    {
      id: "payment",
      icon: CreditCard,
      link: "Payment methods",
      path: "/user/payment",
    },
    {
      id: "reviews",
      icon: Star,
      link: "Reviews",
      path: "/user/reviews",
    },
    {
      id: "settings",
      icon: Settings,
      link: "Settings",
      path: "/user/settings",
    },
  ];

  return (
    <ul className="flex flex-col gap-1">
      {links.map((value) => {
        const isActive = pathname === value.path;
        return (
          <li
            key={value.id}
            className="hover:bg-sidebar-accent overflow-hidden rounded-md transition-colors"
          >
            <Link
              href={value.path}
              className={`flex gap-2.5 items-center p-2.5 transition-colors ${
                isActive
                  ? "text-accent-brand bg-sidebar-accent font-medium"
                  : "text-sidebar-foreground"
              }`}
            >
              <value.icon className="size-5 shrink-0" />
              <span className="hidden md:inline-block text-txt-sm md:text-txt-md">
                {value.link}
              </span>
            </Link>
          </li>
        );
      })}
      <li className="hover:bg-sidebar-accent overflow-hidden rounded-md transition-colors text-red-500">
        <Link
          href="/auth"
          className="flex gap-2.5 items-center p-2.5 text-destructive"
        >
          <LogOut className="size-5 shrink-0" />
          <span className="hidden md:inline-block text-txt-sm md:text-txt-md">
            Logout
          </span>
        </Link>
      </li>
    </ul>
  );
}
