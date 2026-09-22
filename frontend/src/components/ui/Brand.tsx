"use client";

import { Link } from "@/i18n/navigation";

export function Brand() {
  return (
    <Link
      href="/"
      className="inline-block hover:scale-95 transition-transform col-span-3 md:col-span-1 md:justify-self-start font-black text-title-sm md:text-title-md lg:text-title-lg text-center"
    >
      <span>PRI</span>
      <span className="text-accent-brand">M</span>
    </Link>
  );
}
