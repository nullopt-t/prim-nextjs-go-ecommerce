import React from "react";
import Image from "next/image";
import { Link } from "@/i18n/navigation";

interface RecentlyCardProps {
  title: string;
  img?: string;
  slug?: string;
}

export default function RecentlyCard({ title, img, slug }: RecentlyCardProps) {
  const content = (
    <div className="w-24 cursor-pointer group">
      <div className="w-24 h-24 bg-border/40 rounded-xl overflow-hidden border border-border group-hover:border-accent-brand transition-colors relative flex items-center justify-center">
        {img ? (
          <img
            src={img}
            alt={title}
            className="w-full h-full object-cover group-hover:scale-105 transition-transform"
          />
        ) : (
          <div className="w-full h-full bg-secondary/60 flex items-center justify-center text-muted-foreground text-xs font-bold">
            PRIM
          </div>
        )}
      </div>
      <p className="group-hover:text-accent-brand text-[12px] text-muted-foreground mt-2 text-center capitalize line-clamp-2">
        {title}
      </p>
    </div>
  );

  if (slug) {
    return <Link href={`/products/${slug}`}>{content}</Link>;
  }

  return content;
}
