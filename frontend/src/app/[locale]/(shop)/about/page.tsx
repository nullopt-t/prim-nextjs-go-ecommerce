import AboutLayout from "@/features/about/components/layout/aboutLayout";
import { Metadata } from "next";
import { getTranslations } from "next-intl/server";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("meta.shop.about");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function AboutPage() {
  return <AboutLayout />;
}
