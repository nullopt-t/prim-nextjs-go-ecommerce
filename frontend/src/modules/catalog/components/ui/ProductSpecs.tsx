import { useTranslations } from "next-intl";
import { ArrowUpRight, ListChecks, Truck } from "lucide-react";
import { Link } from "@/shared/i18n/navigation";
import type { CategoryAttribute } from "../../types";
import { ProductSection } from "./ProductSection";

interface ProductSpecsProps {
	brand?: string;
	category?: string;
	description?: string;
	/** Attribute values carried by the selected variant. */
	attributes?: Record<string, unknown>;
	/**
	 * The category's own attribute definitions. Their labels are what the
	 * merchant configured, so the spec rows use them instead of a guessed
	 * translation of the raw key.
	 */
	categoryAttributes?: CategoryAttribute[];
}

export function ProductSpecs({
	category,
	brand,
	description,
	attributes,
	categoryAttributes,
}: ProductSpecsProps) {
	const t = useTranslations("products");

  const attributeLabel = (key: string) => {
    // The category's attribute definitions carry the merchant's own labels;
    // prefer them over the guessed key translation.
    const defined = categoryAttributes?.find((attribute) => attribute.key === key)?.label;
    if (defined) return defined;
    const normalized = key.toLowerCase().replace(/[_-]/g, " ");
    if (normalized === "color") return t("actions.color").replace(/[:：]/g, "").trim();
    if (normalized === "storage") return t("specs.storage");
    return normalized;
  };

  // Skip null/empty specs (jsonb attributes may be null) and stringify
  // object/array values instead of rendering "null" / "[object Object]".
  const attributeRows = (attributes ? Object.entries(attributes) : []).filter(
    ([, val]) => val !== null && val !== undefined && String(val).trim() !== "",
  );

  return (
    <div className="flex flex-col gap-5 sm:gap-6">
      <div className="grid grid-cols-1 gap-5 lg:grid-cols-12 sm:gap-6">
        {/* Description Section: only when the catalog has real copy */}
        {description ? (
        <div className="lg:col-span-7">
          <ProductSection title={t("specs.description")}>
            <p
              dir="auto"
              className="whitespace-pre-line text-start text-sm leading-7 text-foreground/90 sm:text-base"
            >
              {description}
            </p>
          </ProductSection>
        </div>
        ) : null}

        {/* Specifications Section */}
        <div className={description ? "lg:col-span-5" : "lg:col-span-12"}>
          <ProductSection
            title={t("specs.specifications")}
            icon={<ListChecks className="size-4 text-accent-brand" aria-hidden="true" />}
          >
            <div className="divide-y divide-border/70 text-sm">
              {brand ? (
                <div className="flex items-center justify-between gap-4 py-3 first:pt-0">
                  <span className="text-muted-foreground">{t("specs.brand")}</span>
                  <span className="text-end font-semibold text-foreground">{brand}</span>
                </div>
              ) : null}
              {category ? (
                <div className="flex items-center justify-between gap-4 py-3">
                  <span className="text-muted-foreground">{t("specs.category")}</span>
                  <span className="text-end font-semibold capitalize text-foreground">
                    {category.replace("-", " ")}
                  </span>
                </div>
              ) : null}
              {attributeRows.map(([key, val]) => (
                <div key={key} className="flex items-center justify-between gap-4 py-3">
                  <span className="capitalize text-muted-foreground">{attributeLabel(key)}</span>
                  <span className="max-w-[60%] text-end font-semibold text-foreground">
                    {typeof val === "object" ? JSON.stringify(val) : String(val)}
                  </span>
                </div>
              ))}
            </div>
          </ProductSection>
        </div>
      </div>

      {/* Shipping & Returns: link to the policy pages so the copy always
          matches the store settings instead of a stale inline summary. */}
      <ProductSection
        title={t("specs.shippingReturns")}
        icon={<Truck className="size-4 text-accent-brand" aria-hidden="true" />}
      >
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2 md:gap-10">
          <div>
            <h3 className="text-sm font-semibold text-foreground">
              {t("specs.shippingInfo")}
            </h3>
            <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
              {t("specs.shippingInfoText")}
            </p>
            <Link
              href="/shipping"
              className="mt-2 inline-flex items-center gap-1 text-sm font-medium text-accent-brand hover:underline"
            >
              {t("specs.shippingPolicyLink")}
              <ArrowUpRight className="size-3.5 rtl:-scale-x-100" aria-hidden="true" />
            </Link>
          </div>

          <div className="border-t border-border pt-6 md:border-s md:border-t-0 md:ps-10 md:pt-0">
            <h3 className="text-sm font-semibold text-foreground">
              {t("specs.returnPolicy")}
            </h3>
            <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
              {t("specs.returnPolicyText")}
            </p>
            <Link
              href="/returns"
              className="mt-2 inline-flex items-center gap-1 text-sm font-medium text-accent-brand hover:underline"
            >
              {t("specs.returnPolicyLink")}
              <ArrowUpRight className="size-3.5 rtl:-scale-x-100" aria-hidden="true" />
            </Link>
          </div>
        </div>
      </ProductSection>
    </div>
  );
}
