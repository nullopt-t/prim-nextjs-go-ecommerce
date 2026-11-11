import { useTranslations } from "next-intl";

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
    <section className="mt-5">
      <h2 className="text-xs font-semibold uppercase tracking-[0.14em] text-muted-foreground">
        {t("tags.title")}
      </h2>
      <ul className="mt-2.5 flex flex-wrap gap-2">
        {items.map((name) => (
          <li
            key={name}
            className="rounded-full border border-border/70 bg-secondary/60 px-2.5 py-1 text-[11px] font-medium text-muted-foreground"
          >
            {name}
          </li>
        ))}
      </ul>
    </section>
  );
}
