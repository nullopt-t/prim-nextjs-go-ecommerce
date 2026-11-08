import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { CompareContent } from "@/modules/compare";

export async function generateMetadata(): Promise<Metadata> {
	const t = await getTranslations("meta.shop.compare");
	return { title: t("title"), description: t("description") };
}

export default function ComparePage() {
	return <CompareContent />;
}
