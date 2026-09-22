import ReviewCard, { ReviewItem } from "@/features/reviews/components/ui/reviewCard";

export default function ReviewGrid({ reviews }: { reviews: ReviewItem[] }) {
  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
      {reviews.map((review) => (
        <ReviewCard key={review.id} reviewDetails={review} />
      ))}
    </div>
  );
}
