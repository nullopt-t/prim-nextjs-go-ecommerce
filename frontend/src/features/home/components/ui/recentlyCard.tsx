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
    <div className="w-28 sm:w-32 cursor-pointer group">
      <div className="w-28 h-28 sm:w-32 sm:h-32 bg-secondary/30 rounded-xl overflow-hidden border border-border group-hover:border-accent-brand group-hover:shadow-md transition-all relative flex items-center justify-center">
        {img ? (
          <img
            src={img}
            alt={title}
            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
          />
        ) : (
          <div className="w-full h-full bg-secondary/60 flex items-center justify-center text-muted-foreground text-xs font-bold">
            PRIM
          </div>
        )}
      </div>
      <p className="group-hover:text-accent-brand text-xs font-medium text-muted-foreground mt-2 text-center capitalize line-clamp-2 transition-colors">
        {title}
      </p>
    </div>
  );

  if (slug) {
    return <Link href={`/products/${slug}`}>{content}</Link>;
  }

  return content;
}
