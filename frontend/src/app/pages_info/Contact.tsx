import MainLayout from "@/components/layouts/mainLayout";
import { useTranslations } from "next-intl";

export function Contact() {
  const t = useTranslations("common.footer");
  return (
    <MainLayout>
      <div className="max-w-4xl mx-auto flex flex-col gap-6">
        <h1 className="text-3xl md:text-4xl font-bold text-foreground">
          {t("contactUs", { defaultMessage: "Contact Us" })}
        </h1>
        <div className="text-muted-foreground flex flex-col gap-4 text-base leading-relaxed">
          <p>We are here to help with any questions, feedback, or support inquiries.</p>
          <div className="mt-4 p-5 bg-secondary/30 rounded-2xl border border-border space-y-2">
            <p className="font-semibold text-foreground">Customer Support:</p>
            <p>Email: support@prim.shop</p>
            <p>Hours: Monday – Saturday, 9:00 AM – 8:00 PM</p>
          </div>
        </div>
      </div>
    </MainLayout>
  );
}
