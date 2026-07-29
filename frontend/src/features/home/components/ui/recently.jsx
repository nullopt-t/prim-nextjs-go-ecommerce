import RecentlyGrid from "@/features/home/components/ui/recentlyGrid";
import { useTranslation } from "react-i18next";

export default function Recently() {
	const { t } = useTranslation("home");

	return (
		<div className="bg-border/25 p-5 m-2.5 mb-0 md:mb-0 lg:mb-0 md:m-5 lg:m-10 border-t border-border">
			<p className="text-muted-foreground font-medium text-sm">
				{t("recentlyViewed.title")}
			</p>
			<RecentlyGrid />
		</div>
	);
}
