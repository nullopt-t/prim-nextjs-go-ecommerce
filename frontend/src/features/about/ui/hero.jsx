import { useTranslation } from "react-i18next";

export default function Hero() {
	const { t } = useTranslation("about");

	return (
		<div className="bg-footer h-72 flex flex-col items-center justify-center">
			<p className="text-center uppercase text-accent-brand text-sm mb-5">
				{t("hero.badge")}
			</p>
			<p className="text-center text-title-sm md:text-title-md lg:text-title-lg text-white font-black mb-2.5">
				{t("hero.title")}
			</p>
			<p className="text-muted-foreground text-center">
				{t("hero.description")}
			</p>
		</div>
	);
}
