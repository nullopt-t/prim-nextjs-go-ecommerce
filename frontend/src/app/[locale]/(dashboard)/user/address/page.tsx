import { Metadata } from "next";
import { AddressesContent } from "@/features/user";
import { getTranslations } from "next-intl/server";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("meta.dashboard.user.addresses");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function AddressPage() {
  return <AddressesContent />;
}
