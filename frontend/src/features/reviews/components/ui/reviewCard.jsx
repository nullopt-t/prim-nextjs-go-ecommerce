import { Text, Stars, CustomButton } from "@/components/ui";
import { useTranslation } from "react-i18next";

export default function ReviewCard({ reviewDetails }) {
	const { t } = useTranslation("reviews");

	return (
		<div className="border border-border rounded-md p-5">
			<Text text={reviewDetails.productName} className="font-medium" />
			<p className="flex items-center gap-2.5">
				<Stars starsNum={reviewDetails.starsNumber} />
				<span className="text-muted-foreground">
					{reviewDetails.starsNumber}
				</span>
			</p>
			<p className="text-muted-foreground mt-2.5 text-txt-sm md:text-txt-md lg:text-txt-lg max-h-48 overflow-auto">
				{reviewDetails.review}
			</p>
			<p className="flex flex-col md:flex-row md:items-center gap-2.5 mt-2.5 mb-5 text-txt-sm md:text-txt-md lg:text-txt-lg">
				<span className="text-foreground">
					{t("reviews.commented")}
				</span>
				<span className="text-muted-foreground">
					{reviewDetails.commentedAt}
				</span>
			</p>
			<div className="flex flex-col md:flex-row gap-2.5">
				<div className="flex-1 bg-primary text-primary-foreground rounded-md hover:bg-accent hover:text-accent-foreground">
					<CustomButton
						text={t("reviews.editReview")}
						onClick={() => {}}
					/>
				</div>
				<div className="flex-1 bg-destructive text-destructive-foreground rounded-md hover:bg-destructive-hover hover:text-destructive-foreground-hover">
					<CustomButton
						text={t("reviews.delete")}
						onClick={() => {}}
					/>
				</div>
			</div>
		</div>
	);
}
