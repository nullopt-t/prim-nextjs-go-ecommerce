import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { Shipping } from "@/modules/content/components/pages/Shipping";

export async function generateMetadata(): Promise<Metadata> {
	const t = await getTranslations("meta.shop.shipping");
	return { title: t("title"), description: t("description") };
}

export default function ShippingPage() {
	return <Shipping />;
}
