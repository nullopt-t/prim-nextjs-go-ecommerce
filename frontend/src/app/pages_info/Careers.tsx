import MainLayout from "@/components/layouts/mainLayout";
import { useTranslations } from "next-intl";

export function Careers() {
  const t = useTranslations("common.footer");
  return (
    <MainLayout>
      <div className="max-w-4xl mx-auto flex flex-col gap-6">
        <h1 className="text-3xl md:text-4xl font-bold text-foreground">
          {t("careers", { defaultMessage: "Careers at PRIM" })}
        </h1>
        <div className="text-muted-foreground flex flex-col gap-4 text-base leading-relaxed">
          <p>Join a fast-growing team redefining modern e-commerce. We are constantly searching for passionate designers, engineers, product builders, and merchant operators.</p>
          <p>Interested in joining? Send your CV and portfolio to careers@prim.shop.</p>
        </div>
      </div>
    </MainLayout>
  );
}
