import { Truck, ShieldCheck } from "lucide-react";
import { useTranslations } from "next-intl";

interface ProductSpecsProps {
  productName: string;
  category: string;
  brand?: string;
  description?: string;
  attributes?: Record<string, any>;
}

export function ProductSpecs({ productName, category, brand, description, attributes }: ProductSpecsProps) {
  const t = useTranslations("products");
  const displayBrand = brand || productName.split(" ")[0] || "PRIM";

  return (
    <div className="mt-20 flex flex-col gap-16">
      {/* Description Section */}
      <section>
        <h2 className="text-2xl font-bold text-foreground mb-6">{t("specs.description")}</h2>
        <div className="prose prose-sm max-w-none text-muted-foreground bg-card border border-border rounded-2xl p-6 md:p-8 shadow-sm">
          {description ? (
            <p className="whitespace-pre-line leading-relaxed text-foreground/90">{description}</p>
          ) : (
            <>
              <p>
                {t("specs.fallbackIntro", { productName })}
              </p>
              <ul className="mt-4 space-y-2">
                <li>{t("specs.fallbackFeature1")}</li>
                <li>{t("specs.fallbackFeature2")}</li>
                <li>{t("specs.fallbackFeature3")}</li>
                <li>{t("specs.fallbackFeature4")}</li>
              </ul>
            </>
          )}
        </div>
      </section>

      {/* Specifications Section */}
      <section>
        <h2 className="text-2xl font-bold text-foreground mb-6">{t("specs.specifications")}</h2>
        <div className="bg-card border border-border rounded-2xl p-6 md:p-8 shadow-sm">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-12 gap-y-4 text-sm">
            <div className="flex justify-between py-3 border-b border-border/50">
              <span className="text-muted-foreground">{t("specs.brand")}</span>
              <span className="font-medium text-foreground">{displayBrand}</span>
            </div>
            <div className="flex justify-between py-3 border-b border-border/50">
              <span className="text-muted-foreground">{t("specs.category")}</span>
              <span className="font-medium text-foreground capitalize">{category.replace("-", " ")}</span>
            </div>
            {attributes &&
              Object.entries(attributes).map(([key, val]) => (
                <div key={key} className="flex justify-between py-3 border-b border-border/50">
                  <span className="text-muted-foreground capitalize">{key}</span>
                  <span className="font-medium text-foreground">{String(val)}</span>
                </div>
              ))}
            <div className="flex justify-between py-3 border-b border-border/50">
              <span className="text-muted-foreground">{t("specs.condition")}</span>
              <span className="font-medium text-foreground">{t("specs.brandNew")}</span>
            </div>
            <div className="flex justify-between py-3 border-b border-border/50">
              <span className="text-muted-foreground">{t("specs.warranty")}</span>
              <span className="font-medium text-foreground">{t("specs.warrantyValue")}</span>
            </div>
          </div>
        </div>
      </section>

      {/* Shipping & Returns Section */}
      <section>
        <h2 className="text-2xl font-bold text-foreground mb-6">{t("specs.shippingReturns")}</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="bg-card border border-border rounded-2xl p-6 flex flex-col gap-3 shadow-sm hover:border-accent-brand/50 transition-colors">
            <div className="w-10 h-10 rounded-full bg-secondary/50 flex items-center justify-center text-foreground">
              <Truck className="size-5" />
            </div>
            <h4 className="text-foreground font-bold text-base">{t("specs.shippingInfo")}</h4>
            <p className="text-sm text-muted-foreground leading-relaxed">
              {t("specs.shippingInfoText")}
            </p>
          </div>

          <div className="bg-card border border-border rounded-2xl p-6 flex flex-col gap-3 shadow-sm hover:border-accent-brand/50 transition-colors">
            <div className="w-10 h-10 rounded-full bg-secondary/50 flex items-center justify-center text-foreground">
              <ShieldCheck className="size-5" />
            </div>
            <h4 className="text-foreground font-bold text-base">{t("specs.returnPolicy")}</h4>
            <p className="text-sm text-muted-foreground leading-relaxed">
              {t("specs.returnPolicyText")}
            </p>
          </div>
        </div>
      </section>
    </div>
  );
}
