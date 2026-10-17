import { Metadata } from "next";
import OrderDetailsContent from "@/features/user/components/ui/orderDetailsContent";
import { getTranslations } from "next-intl/server";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("meta.dashboard.user.orderDetails");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function OrderDetailsPage() {
  return <OrderDetailsContent />;
}
