import { ReviewsLayout, ReviewItem } from "@/features/reviews";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Your Reviews | PRIM",
  description: "View and manage your product reviews",
};

const sampleReviews: ReviewItem[] = [
  {
    id: 1,
    productName: "Sony WH-1000XM5 Wireless Headphones",
    starsNumber: 5,
    review: "The sound quality is sublime, and the active noise cancellation works like magic.",
    commentedAt: "2026-07-20",
  },
  {
    id: 2,
    productName: "Apple AirPods Max",
    starsNumber: 4,
    review: "Excellent build quality and transparency mode. A bit on the heavier side.",
    commentedAt: "2026-07-18",
  },
];

export default function ReviewsPage() {
  return <ReviewsLayout reviews={sampleReviews} />;
}
