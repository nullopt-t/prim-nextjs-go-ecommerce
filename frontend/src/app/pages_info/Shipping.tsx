import MainLayout from "@/components/layouts/mainLayout";
import { useTranslations } from "next-intl";

export function Shipping() {
  const t = useTranslations("shop");
  return (
    <MainLayout>
      <div className="w-full flex flex-col gap-6 py-4 sm:py-6">
        <h1 className="text-3xl md:text-4xl font-bold text-foreground">
          {t("pages.shipping.title")}
        </h1>
        <div className="text-muted-foreground flex flex-col gap-4 text-base leading-relaxed">
          <p>{t("pages.shipping.intro")}</p>
          <h3 className="text-xl font-semibold text-foreground mt-4">{t("pages.shipping.standardTitle")}</h3>
          <p>{t("pages.shipping.standardDescription")}</p>
          <h3 className="text-xl font-semibold text-foreground mt-4">{t("pages.shipping.expressTitle")}</h3>
          <p>{t("pages.shipping.expressDescription")}</p>
        </div>
      </div>
    </MainLayout>
  );
}
