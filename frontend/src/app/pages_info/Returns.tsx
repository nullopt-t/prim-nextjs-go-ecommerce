import MainLayout from "@/components/layouts/mainLayout";
import { useTranslations } from "next-intl";

export function Returns() {
  const t = useTranslations("common.footer");
  return (
    <MainLayout>
      <div className="max-w-4xl mx-auto flex flex-col gap-6">
        <h1 className="text-3xl md:text-4xl font-bold text-foreground">
          {t("returns", { defaultMessage: "Returns & Exchanges" })}
        </h1>
        <div className="text-muted-foreground flex flex-col gap-4 text-base leading-relaxed">
          <p>We want you to be completely satisfied with your purchase. If you are not happy with your order, we are here to help.</p>
          <h3 className="text-xl font-semibold text-foreground mt-4">Return Policy</h3>
          <p>You have 30 calendar days to return an item from the date you received it. To be eligible for a return, your item must be unused and in the same condition that you received it. Your item must be in the original packaging.</p>
          <h3 className="text-xl font-semibold text-foreground mt-4">Refunds</h3>
          <p>Once we receive your item, we will inspect it and notify you that we have received your returned item. We will immediately notify you on the status of your refund after inspecting the item. If your return is approved, we will initiate a refund to your credit card (or original method of payment).</p>
        </div>
      </div>
    </MainLayout>
  );
}
