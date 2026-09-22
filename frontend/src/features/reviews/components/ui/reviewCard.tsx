"use client";

import { Text, Stars, CustomButton } from "@/components/ui";
import { useTranslations } from "next-intl";

export interface ReviewItem {
  id: string | number;
  productName: string;
  starsNumber: number;
  review: string;
  commentedAt: string;
}

export default function ReviewCard({ reviewDetails }: { reviewDetails: ReviewItem }) {
  const t = useTranslations("reviews");

  return (
    <div className="border border-border rounded-md p-5 bg-card">
      <Text text={reviewDetails.productName} className="font-medium" />
      <div className="flex items-center gap-2.5 my-2">
        <Stars starsNum={reviewDetails.starsNumber} />
        <span className="text-muted-foreground text-xs">
          {reviewDetails.starsNumber}
        </span>
      </div>
      <p className="text-muted-foreground mt-2.5 text-txt-sm md:text-txt-md max-h-48 overflow-auto leading-relaxed">
        {reviewDetails.review}
      </p>
      <div className="flex flex-col md:flex-row md:items-center gap-2 mt-3 mb-5 text-txt-sm">
        <span className="text-foreground font-medium">
          {t("reviews.commented")}
        </span>
        <span className="text-muted-foreground">
          {reviewDetails.commentedAt}
        </span>
      </div>
      <div className="flex flex-col sm:flex-row gap-2.5">
        <div className="flex-1 bg-primary text-primary-foreground rounded-md hover:opacity-90">
          <CustomButton text={t("reviews.editReview")} onClick={() => {}} />
        </div>
        <div className="flex-1 bg-destructive text-destructive-foreground rounded-md hover:opacity-90">
          <CustomButton text={t("reviews.delete")} onClick={() => {}} />
        </div>
      </div>
    </div>
  );
}
