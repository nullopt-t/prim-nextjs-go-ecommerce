"use client";

import { Link } from "@/shared/i18n/navigation";
import { usePublicSettings } from "@/modules/settings/model/publicSettingsContext";

export function Brand() {
  const { settings } = usePublicSettings();
  const storeName = settings?.storeName?.trim() || "PRIM";
  const [primary, ...rest] = storeName.split(/\s+/);
  const secondary = rest.join(" ");
  const isPrim = primary.toUpperCase() === "PRIM";

  return (
    <Link
      href="/"
      aria-label={storeName}
      className="inline-flex items-center gap-1.5 group select-none cursor-pointer max-w-[11rem] sm:max-w-[18rem]"
    >
      <span dir="ltr" className="font-black text-2xl tracking-tighter text-foreground group-hover:opacity-90 transition-opacity whitespace-nowrap truncate">
        {isPrim ? (
          <>
            PRI<span className="text-accent-brand">M</span>
          </>
        ) : (
          primary
        )}
        {secondary && (
          <span className="ms-1 text-[0.6em] font-bold tracking-normal text-muted-foreground">
            {secondary}
          </span>
        )}
      </span>
    </Link>
  );
}
