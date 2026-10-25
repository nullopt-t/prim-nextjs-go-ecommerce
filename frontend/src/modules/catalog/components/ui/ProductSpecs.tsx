import type { ReactNode } from "react";
import { useTranslations } from "next-intl";

interface ProductSpecsProps {
	productName: string;
	category?: string;
	brand?: string;
	description?: string;
	attributes?: Record<string, any>;
	condition?: string;
	warranty?: string;
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
	productName,
	category,
	brand,
	description,
	attributes,
	condition,
	warranty,
}: ProductSpecsProps) {
	const t = useTranslations("products");

  const attributeLabel = (key: string) => {
    const normalized = key.toLowerCase().replace(/[_-]/g, " ");
    if (normalized === "color") return t("actions.color").replace(/[:：]/g, "").trim();
    if (normalized === "storage") return t("specs.storage");
    return normalized;
  };

  const attributeRows = attributes ? Object.entries(attributes) : [];

  return (
    <div className="mt-10 flex flex-col gap-8 sm:mt-14 sm:gap-10">
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
        {/* Description Section */}
        <section className="lg:col-span-7">
          <SectionHeading>{t("specs.description")}</SectionHeading>
          <div className="h-full rounded-2xl border border-border bg-card p-6 shadow-sm sm:p-8">
            {description ? (
              <p
                dir="auto"
                className="whitespace-pre-line text-start text-sm leading-7 text-foreground/90 sm:text-base"
              >
                {description}
              </p>
            ) : (
              <div className="space-y-4 text-sm leading-7 text-muted-foreground sm:text-base">
                <p>{t("specs.fallbackIntro", { productName })}</p>
                <ul className="grid gap-2.5 sm:grid-cols-2">
                  {[
                    t("specs.fallbackFeature1"),
                    t("specs.fallbackFeature2"),
                    t("specs.fallbackFeature3"),
                    t("specs.fallbackFeature4"),
                  ].map((feature) => (
                    <li key={feature} className="flex items-start gap-2">
                      <span className="mt-2.5 size-1 shrink-0 rounded-full bg-muted-foreground/50" />
                      <span>{feature}</span>
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>
        </section>

        {/* Specifications Section */}
        <section className="lg:col-span-5">
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
                    {String(val)}
                  </span>
                </div>
              ))}
              {condition ? (
                <div className="flex items-center justify-between gap-4 py-3">
                  <span className="text-muted-foreground">{t("specs.condition")}</span>
                  <span className="text-end font-semibold text-foreground">{condition}</span>
                </div>
              ) : null}
              {warranty ? (
                <div className="flex items-center justify-between gap-4 py-3 last:pb-0">
                  <span className="text-muted-foreground">{t("specs.warranty")}</span>
                  <span className="text-end font-semibold text-foreground">{warranty}</span>
                </div>
              ) : null}
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
