import MainLayout from "@/components/layouts/mainLayout";
import { useTranslations } from "next-intl";

export function Careers() {
  const t = useTranslations("shop");
  return (
    <MainLayout>
      <div className="w-full flex flex-col gap-6 py-4 sm:py-6">
        <h1 className="text-3xl md:text-4xl font-bold text-foreground">
          {t("pages.careers.title")}
        </h1>
        <div className="text-muted-foreground flex flex-col gap-4 text-base leading-relaxed">
          <p>{t("pages.careers.description")}</p>
          <p>{t("pages.careers.apply")}</p>
        </div>
      </div>
    </MainLayout>
  );
}
