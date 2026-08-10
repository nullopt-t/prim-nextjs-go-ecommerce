import ReviewCard from "@/features/reviews/components/ui/reviewCard";

export default function ReviewGrid({ reviews }) {
	return (
		<div className="grid grid-cols-1 lg:grid-cols-2 gap-2.5">
			{reviews.map((review) => (
				<ReviewCard key={review.id} reviewDetails={review} />
			))}
		</div>
	);
}
