import { useTranslations } from "next-intl";
import { Sparkles } from "lucide-react";
import { ProductSection } from "./ProductSection";

interface ProductHighlightsProps {
  highlights?: string[] | null;
}

/**
 * The merchant's own bullet points for the product. These come from the
 * database (`products.highlights`), not from a fixed list.
 */
export function ProductHighlights({ highlights }: ProductHighlightsProps) {
  const t = useTranslations("products");

  const items = Array.isArray(highlights)
    ? highlights.map((item) => String(item ?? "").trim()).filter(Boolean)
    : [];

  // A missing or empty highlights list should not leave an empty section behind.
  if (items.length === 0) return null;

  return (
    <ProductSection title={t("highlights.title")} icon={<Sparkles className="size-4 text-accent-brand" aria-hidden="true" />}>
      <ul className="divide-y divide-border/60">
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
    </ProductSection>
  );
}
