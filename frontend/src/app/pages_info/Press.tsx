import MainLayout from "@/components/layouts/mainLayout";
import { useTranslations } from "next-intl";

export function Press() {
  const t = useTranslations("common.footer");
  return (
    <MainLayout>
      <div className="max-w-4xl mx-auto flex flex-col gap-6">
        <h1 className="text-3xl md:text-4xl font-bold text-foreground">
          {t("press", { defaultMessage: "Press & News" })}
        </h1>
        <div className="text-muted-foreground flex flex-col gap-4 text-base leading-relaxed">
          <p>Read about our latest product releases, company updates, and platform milestones.</p>
          <p>For press inquiries and interview requests: press@prim.shop</p>
        </div>
      </div>
    </MainLayout>
  );
}
