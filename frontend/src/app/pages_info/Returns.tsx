import MainLayout from "@/components/layouts/mainLayout";
import { useTranslations } from "next-intl";

export function Returns() {
  const t = useTranslations("shop");
  return (
    <MainLayout>
      <div className="w-full flex flex-col gap-6 py-4 sm:py-6">
        <h1 className="text-3xl md:text-4xl font-bold text-foreground">
          {t("pages.returns.title")}
        </h1>
        <div className="text-muted-foreground flex flex-col gap-4 text-base leading-relaxed">
          <p>{t("pages.returns.intro")}</p>
          <h3 className="text-xl font-semibold text-foreground mt-4">{t("pages.returns.policyTitle")}</h3>
          <p>{t("pages.returns.policyDescription")}</p>
          <h3 className="text-xl font-semibold text-foreground mt-4">{t("pages.returns.refundsTitle")}</h3>
          <p>{t("pages.returns.refundsDescription")}</p>
        </div>
      </div>
    </MainLayout>
  );
}
