import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { Returns } from "@/modules/content/components/pages/Returns";

export async function generateMetadata(): Promise<Metadata> {
	const t = await getTranslations("meta.shop.returns");
	return { title: t("title"), description: t("description") };
}

export default function ReturnsPage() {
	return <Returns />;
}
