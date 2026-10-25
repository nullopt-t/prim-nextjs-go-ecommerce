import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { Categories } from "@/modules/catalog/components/categories";

export async function generateMetadata(): Promise<Metadata> {
	const t = await getTranslations("meta.shop.categories");
	return { title: t("title"), description: t("description") };
}

export default function CategoriesPage() {
	return <Categories />;
}
