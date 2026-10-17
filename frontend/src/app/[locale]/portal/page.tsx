import { PortalManager } from "@/features/portal/portalManager";
import { Metadata } from "next";
import { getTranslations } from "next-intl/server";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("meta.portal");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function PortalPage() {
  return <PortalManager />;
}
