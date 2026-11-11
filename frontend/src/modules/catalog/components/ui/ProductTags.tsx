import { useTranslations } from "next-intl";
import { Tag } from "lucide-react";
import { ProductSection } from "./ProductSection";

interface ProductTagsProps {
  tags?: Array<{ id?: string; name?: string }>;
}

/**
 * Taxonomy tags assigned to the product. The backend sends them on the
 * product details payload; they were previously dropped by the normalizer.
 */
export function ProductTags({ tags }: ProductTagsProps) {
  const t = useTranslations("products");

  const items = (tags ?? [])
    .map((tag) => String(tag?.name ?? "").trim())
    .filter((name) => name !== "");

  if (items.length === 0) return null;

  return (
    <ProductSection title={t("tags.title")} icon={<Tag className="size-4 text-accent-brand" aria-hidden="true" />}>
      <ul className="flex flex-wrap gap-2">
        {items.map((name) => (
          <li
            key={name}
            className="rounded-full border border-border/70 bg-secondary/60 px-2.5 py-1 text-[11px] font-medium text-muted-foreground"
          >
            {name}
          </li>
        ))}
      </ul>
    </ProductSection>
  );
}
