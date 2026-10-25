import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { Press } from "@/modules/content/components/pages/Press";

export async function generateMetadata(): Promise<Metadata> {
	const t = await getTranslations("meta.shop.press");
	return { title: t("title"), description: t("description") };
}

export default function PressPage() {
	return <Press />;
}
