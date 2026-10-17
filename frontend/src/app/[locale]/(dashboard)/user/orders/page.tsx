import { Metadata } from "next";
import { OrdersContent } from "@/features/user";
import { getTranslations } from "next-intl/server";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("meta.dashboard.user.orders");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function OrdersPage() {
  return <OrdersContent />;
}
