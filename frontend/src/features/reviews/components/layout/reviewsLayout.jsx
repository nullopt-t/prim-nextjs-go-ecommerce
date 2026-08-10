import { Title } from "@/components/ui";
import ReviewGrid from "@/features/reviews/components/ui/reviewGrid";

export default function ReviewsLayout({ reviews }) {
	return (
		<div className="">
			<Title
				title="Your Reviews"
				subtitle="Reviews you've written about your purchases."
			/>
			<ReviewGrid reviews={reviews} />
		</div>
	);
}
