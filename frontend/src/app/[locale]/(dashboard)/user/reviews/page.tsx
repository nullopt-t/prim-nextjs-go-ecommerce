import { Metadata } from "next";
import { ReviewsContent } from "@/features/user";
import { getTranslations } from "next-intl/server";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("meta.dashboard.user.reviews");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function ReviewsPage() {
  return <ReviewsContent />;
}
