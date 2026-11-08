import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import OrderLookupContent from "@/modules/orders/components/ui/OrderLookupContent";

export async function generateMetadata(): Promise<Metadata> {
	const t = await getTranslations("meta.shop.orderLookup");
	return { title: t("title"), description: t("description") };
}

export default function OrderLookupPage() {
	return <OrderLookupContent />;
}
