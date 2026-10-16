import MainLayout from "@/components/layouts/mainLayout";
import { useTranslations } from "next-intl";

export function Faqs() {
  const t = useTranslations("shop");
  return (
    <MainLayout>
      <div className="w-full flex flex-col gap-6 py-4 sm:py-6">
        <h1 className="text-3xl md:text-4xl font-bold text-foreground">
          {t("pages.faqs.title")}
        </h1>
        <div className="text-muted-foreground flex flex-col gap-6 text-base leading-relaxed">
          <div>
            <h3 className="text-xl font-semibold text-foreground mb-2">{t("pages.faqs.trackOrder.question")}</h3>
            <p>{t("pages.faqs.trackOrder.answer")}</p>
          </div>
          <div>
            <h3 className="text-xl font-semibold text-foreground mb-2">{t("pages.faqs.changeOrder.question")}</h3>
            <p>{t("pages.faqs.changeOrder.answer")}</p>
          </div>
          <div>
            <h3 className="text-xl font-semibold text-foreground mb-2">{t("pages.faqs.payment.question")}</h3>
            <p>{t("pages.faqs.payment.answer")}</p>
          </div>
        </div>
      </div>
    </MainLayout>
  );
}
