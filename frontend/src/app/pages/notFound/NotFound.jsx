import { Title } from "@/components/ui";
import { useTranslation } from "react-i18next";
import NotFoundEyes from "@/assets/imgs/placeholders/not-found-eyes.png";
import NotFoundActions from "@/app/pages/notFound/NotFoundActions";

export function NotFound() {
	const { t } = useTranslation("notFound");

	return (
		<div className="p-2.5 md:p-5">
			<div className="flex flex-col items-center">
				<Title title={t("title")} />
				<p className="text-center text-txt-sm md:text-txt-md lg:text-txt-lg">
					{t("subtitle")}
				</p>
				<div className="w-64 md:w-80 lg:w-96">
					<img
						src={NotFoundEyes}
						alt="Not found eyes"
						className="object-center object-cover"
					/>
				</div>
				<NotFoundActions />
			</div>
		</div>
	);
}
