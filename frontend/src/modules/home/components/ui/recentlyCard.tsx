import React from "react";
import { X } from "lucide-react";
import { Link } from "@/shared/i18n/navigation";
import { CatalogImage } from "@/shared/ui/catalogImage";
import { useTranslations } from "next-intl";

interface RecentlyCardProps {
  title: string;
  img?: string;
  slug?: string;
  onRemove?: () => void;
}

export default function RecentlyCard({ title, img, slug, onRemove }: RecentlyCardProps) {
  const t = useTranslations("home");
  const content = (
    <div className="w-28 sm:w-32 cursor-pointer group relative">
      <div className="w-28 h-28 sm:w-32 sm:h-32 bg-secondary/30 rounded-xl overflow-hidden border border-border group-hover:border-accent-brand group-hover:shadow-md transition-all relative flex items-center justify-center">
        {img ? (
          <CatalogImage
            src={img}
            alt={title}
            fill
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
      {onRemove && (
        <button
          type="button"
          onClick={(e) => {
            e.preventDefault();
            e.stopPropagation();
            onRemove();
          }}
          aria-label={t("recentlyViewed.remove", { name: title })}
          className="absolute -top-1.5 -end-1.5 flex size-5 items-center justify-center rounded-full bg-card text-muted-foreground shadow-xs border border-border hover:text-destructive cursor-pointer"
        >
          <X className="size-3" />
        </button>
      )}
    </div>
  );

  if (slug) {
    return <Link href={`/products/${slug}`}>{content}</Link>;
  }

  return content;
}
