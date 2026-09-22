import MainLayout from "@/components/layouts/mainLayout";
import { useTranslations } from "next-intl";

export function Faqs() {
  const t = useTranslations("common.footer");
  return (
    <MainLayout>
      <div className="max-w-4xl mx-auto flex flex-col gap-6">
        <h1 className="text-3xl md:text-4xl font-bold text-foreground">
          {t("faqs", { defaultMessage: "Frequently Asked Questions" })}
        </h1>
        <div className="text-muted-foreground flex flex-col gap-6 text-base leading-relaxed">
          <div>
            <h3 className="text-xl font-semibold text-foreground mb-2">How do I track my order?</h3>
            <p>Once your order has shipped, you will receive an email containing a tracking number and link to track your package in real-time.</p>
          </div>
          <div>
            <h3 className="text-xl font-semibold text-foreground mb-2">Can I change or cancel my order?</h3>
            <p>Orders are processed quickly to ensure fast delivery. If you need to make changes, please visit your account dashboard or contact customer support immediately.</p>
          </div>
          <div>
            <h3 className="text-xl font-semibold text-foreground mb-2">What payment methods do you accept?</h3>
            <p>We accept major credit cards including Visa, MasterCard, American Express, as well as digital payment options.</p>
          </div>
        </div>
      </div>
    </MainLayout>
  );
}
