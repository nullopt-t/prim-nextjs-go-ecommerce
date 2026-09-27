"use client";

import { Link } from "@/i18n/navigation";

export function Brand() {
  return (
    <Link
      href="/"
      className="inline-flex items-center gap-1.5 group select-none cursor-pointer"
    >
      <span dir="ltr" className="font-black text-2xl tracking-tighter text-foreground group-hover:opacity-90 transition-opacity">
        PRI<span className="text-accent-brand">M</span>
      </span>
    </Link>
  );
}
