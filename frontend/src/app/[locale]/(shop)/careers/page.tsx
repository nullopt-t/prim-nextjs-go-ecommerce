import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { Careers } from "@/modules/content/components/pages/Careers";

export async function generateMetadata(): Promise<Metadata> {
	const t = await getTranslations("meta.shop.careers");
	return { title: t("title"), description: t("description") };
}

export default function CareersPage() {
	return <Careers />;
}
