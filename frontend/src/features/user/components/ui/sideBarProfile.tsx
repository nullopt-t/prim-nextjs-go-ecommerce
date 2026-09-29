"use client";

import { useAuthContext } from "@/context/AuthContext";

export default function UserSideBarProfile() {
  const { user } = useAuthContext();

  const displayName = user?.name || "Guest User";
  const displayEmail = user?.email || "";
  const initial = displayName.charAt(0).toUpperCase();

  return (
    <div className="flex items-center gap-3 px-4 py-4 border-b border-border">
      {/* Avatar */}
      <div className="size-10 rounded-full bg-accent-brand/10 text-accent-brand font-bold text-base flex items-center justify-center shrink-0">
        {initial}
      </div>

      {/* Info */}
      <div className="flex-1 min-w-0">
        <p className="font-semibold text-sm text-foreground truncate leading-tight">
          {displayName}
        </p>
        {displayEmail && (
          <p className="text-xs text-muted-foreground truncate leading-tight mt-0.5">
            {displayEmail}
          </p>
        )}
        <span className="text-[10px] text-muted-foreground font-medium mt-0.5 inline-block">
          Customer
        </span>
      </div>
    </div>
  );
}
