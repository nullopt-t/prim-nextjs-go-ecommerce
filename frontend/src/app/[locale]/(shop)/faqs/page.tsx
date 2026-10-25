import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { Faqs } from "@/modules/content/components/pages/Faqs";

export async function generateMetadata(): Promise<Metadata> {
	const t = await getTranslations("meta.shop.faqs");
	return { title: t("title"), description: t("description") };
}

export default function FaqsPage() {
	return <Faqs />;
}
