import { Title } from "@/components/ui";
import ReviewGrid from "@/features/reviews/components/ui/reviewGrid";
import { useTranslation } from "react-i18next";

export default function ReviewsLayout({ reviews }) {
	const { t } = useTranslation("reviews");
	return (
		<div className="">
			<Title
				title={t("reviews.title")}
				subtitle={t("reviews.description")}
			/>
			<ReviewGrid reviews={reviews} />
		</div>
	);
}
