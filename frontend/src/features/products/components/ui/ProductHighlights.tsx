import { useTranslations } from "next-intl";

interface ProductHighlightsProps {
  highlights?: string[] | null;
}

export function ProductHighlights({ highlights }: ProductHighlightsProps) {
  const t = useTranslations("products");

  const items = Array.isArray(highlights)
    ? highlights.map((item) => String(item ?? "").trim()).filter(Boolean)
    : [];

  // A missing or empty highlights list should not leave an empty section behind.
  if (items.length === 0) return null;

  return (
    <section className="border-t border-border pt-5">
      <h2 className="text-xs font-semibold uppercase tracking-[0.14em] text-muted-foreground">
        {t("highlights.title")}
      </h2>

      <ul className="mt-3 divide-y divide-border/60">
        {items.map((item, index) => (
          <li key={`${item}-${index}`} className="py-2.5 first:pt-0 last:pb-0">
            <span
              dir="auto"
              className="block text-start text-sm leading-relaxed text-foreground/90"
            >
              {item}
            </span>
          </li>
        ))}
      </ul>
    </section>
  );
}
