"use client";

import { Link } from "@/i18n/navigation";

export function Brand() {
  return (
    <Link
      href="/"
      className="inline-flex items-center hover:scale-95 transition-transform font-black text-2xl md:text-title-md tracking-tight"
    >
      <span>PRI</span>
      <span className="text-accent-brand">M</span>
    </Link>
  );
}
