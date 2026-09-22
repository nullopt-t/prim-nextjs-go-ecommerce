import MainLayout from "@/components/layouts/mainLayout";
import { useTranslations } from "next-intl";

export function Shipping() {
  const t = useTranslations("common.footer");
  return (
    <MainLayout>
      <div className="max-w-4xl mx-auto flex flex-col gap-6">
        <h1 className="text-3xl md:text-4xl font-bold text-foreground">
          {t("shipping", { defaultMessage: "Shipping Information" })}
        </h1>
        <div className="text-muted-foreground flex flex-col gap-4 text-base leading-relaxed">
          <p>We offer worldwide shipping with various delivery options to ensure you receive your order as quickly and safely as possible.</p>
          <h3 className="text-xl font-semibold text-foreground mt-4">Standard Shipping</h3>
          <p>Standard shipping usually takes 5-7 business days within the contiguous United States. International shipping times may vary based on destination and customs processing.</p>
          <h3 className="text-xl font-semibold text-foreground mt-4">Express Shipping</h3>
          <p>Need it faster? We offer express shipping options at checkout. Express orders are prioritized and typically arrive within 2-3 business days.</p>
        </div>
      </div>
    </MainLayout>
  );
}
