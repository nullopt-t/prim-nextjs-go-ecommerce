import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { Contact } from "@/modules/content/components/pages/Contact";

export async function generateMetadata(): Promise<Metadata> {
	const t = await getTranslations("meta.shop.contact");
	return { title: t("title"), description: t("description") };
}

export default function ContactPage() {
	return <Contact />;
}
