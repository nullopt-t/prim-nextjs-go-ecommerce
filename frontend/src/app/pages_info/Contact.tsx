import MainLayout from "@/components/layouts/mainLayout";
import { useTranslations } from "next-intl";

export function Contact() {
  const t = useTranslations("shop");
  return (
    <MainLayout>
      <div className="w-full flex flex-col gap-6 py-4 sm:py-6">
        <h1 className="text-3xl md:text-4xl font-bold text-foreground">
          {t("pages.contact.title")}
        </h1>
        <div className="text-muted-foreground flex flex-col gap-4 text-base leading-relaxed">
          <p>{t("pages.contact.intro")}</p>
          <div className="mt-4 p-5 bg-secondary/30 rounded-2xl border border-border space-y-2">
            <p className="font-semibold text-foreground">{t("pages.contact.supportLabel")}</p>
            <p>{t("pages.contact.email")}</p>
            <p>{t("pages.contact.hours")}</p>
          </div>
        </div>
      </div>
    </MainLayout>
  );
}
