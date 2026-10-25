import { OverviewContent } from "@/modules/account";
import { Metadata } from "next";
import { getTranslations } from "next-intl/server";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("meta.dashboard.user.overview");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function OverviewPage() {
  return <OverviewContent />;
}
