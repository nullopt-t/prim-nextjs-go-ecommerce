import type { ReactNode } from "react";
import { useTranslations } from "next-intl";
import type { CategoryAttribute } from "../../types";

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

function SectionHeading({ children }: { children: ReactNode }) {
  return (
    <div className="mb-4 border-b border-border pb-2.5">
      <h2 className="text-base font-semibold text-foreground sm:text-lg">
        {children}
      </h2>
    </div>
  );
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
    <div className="mt-10 flex flex-col gap-8 sm:mt-14 sm:gap-10">
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
        {/* Description Section: only when the catalog has real copy */}
        {description ? (
        <section className="lg:col-span-7">
          <SectionHeading>{t("specs.description")}</SectionHeading>
          <div className="h-full rounded-2xl border border-border bg-card p-6 shadow-sm sm:p-8">
            <p
              dir="auto"
              className="whitespace-pre-line text-start text-sm leading-7 text-foreground/90 sm:text-base"
            >
              {description}
            </p>
          </div>
        </section>
        ) : null}

        {/* Specifications Section */}
        <section className={description ? "lg:col-span-5" : "lg:col-span-12"}>
          <SectionHeading>{t("specs.specifications")}</SectionHeading>
          <div className="h-full rounded-2xl border border-border bg-card p-5 shadow-sm sm:p-6">
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
          </div>
        </section>
      </div>

      {/* Shipping & Returns Section */}
      <section>
        <SectionHeading>{t("specs.shippingReturns")}</SectionHeading>
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2 md:gap-10">
          <div>
            <h3 className="text-sm font-semibold text-foreground">
              {t("specs.shippingInfo")}
            </h3>
            <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
              {t("specs.shippingInfoText")}
            </p>
          </div>

          <div className="border-t border-border pt-6 md:border-s md:border-t-0 md:ps-10 md:pt-0">
            <h3 className="text-sm font-semibold text-foreground">
              {t("specs.returnPolicy")}
            </h3>
            <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
              {t("specs.returnPolicyText")}
            </p>
          </div>
        </div>
      </section>
    </div>
  );
}
