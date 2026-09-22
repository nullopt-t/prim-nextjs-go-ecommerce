"use client";

import { Title } from "@/components/ui/title";
import ReviewGrid from "@/features/reviews/components/ui/reviewGrid";
import { ReviewItem } from "@/features/reviews/components/ui/reviewCard";
import { useTranslations } from "next-intl";

export default function ReviewsLayout({ reviews }: { reviews: ReviewItem[] }) {
  const t = useTranslations("reviews");

  return (
    <div>
      <Title
        title={t("reviews.title")}
        subtitle={t("reviews.description")}
      />
      <ReviewGrid reviews={reviews} />
    </div>
  );
}
