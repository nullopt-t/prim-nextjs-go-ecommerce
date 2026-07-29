import { HiMiniArrowLongRight } from "react-icons/hi2";
import { useTranslation } from "react-i18next";

export default function SectionTitle({ title }) {
	const { t } = useTranslation("home");

	return (
		<div className="flex justify-between">
			<p className="font-medium text-title-sm md:text-title-md text-foreground">
				{t(title)}
			</p>
			<p className="flex gap-5 items-center text-txt-sm md:text-txt-md lg:text-txt-lg cursor-pointer text-accent-brand hover:underline hover:underline-offset-4">
				<span className="">{t("categories.seeAll")}</span>
				<span className="">
					<HiMiniArrowLongRight />
				</span>
			</p>
		</div>
	);
}
