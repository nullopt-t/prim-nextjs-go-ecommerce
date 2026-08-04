import { Title } from "@/components/ui";
import { useTranslation } from "react-i18next";
import NotFoundEyes from "@/assets/imgs/placeholders/eyes.svg?react";
import NotFoundActions from "@/app/pages/notFound/NotFoundActions";

export function NotFound() {
	const { t } = useTranslation("notFound");

	return (
		<div className="p-2.5 md:p-5">
			<div className="flex flex-col items-center">
				<Title title={t("title")} />
				<p className="text-center text-foreground text-txt-sm md:text-txt-md lg:text-txt-lg mb-5">
					{t("subtitle")}
				</p>
				<div className="w-64 md:w-80 lg:w-96">
					<NotFoundEyes className="w-full h-auto text-foreground" />
				</div>
				<NotFoundActions />
			</div>
		</div>
	);
}
