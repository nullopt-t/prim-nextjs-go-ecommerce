"use client";

import { useAuthContext } from "@/context/AuthContext";

export default function UserSideBarProfile() {
  const { user } = useAuthContext();

  const displayName = user?.name || "Mohamed Mahmoud";
  const displayEmail = user?.email || "prim@store.com";

  return (
    <div className="p-2.5 hidden md:block border-b border-sidebar-border mb-3">
      <div className="flex flex-col">
        <span className="font-medium text-sidebar-foreground truncate">{displayName}</span>
        <span className="text-muted-foreground text-xs truncate">{displayEmail}</span>
      </div>
    </div>
  );
}
