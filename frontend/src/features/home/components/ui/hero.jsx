import Button from "@/components/ui/button";
import { useTranslation } from "react-i18next";

const currentDate = new Date().getFullYear();

export default function Hero() {
	const { t } = useTranslation("home");

	return (
		<div className="p-15">
			<p className="text-accent-brand mb-5">
				{t("hero.badge", { year: currentDate })}
			</p>
			<p className="mb-2.5 text-white font-black text-title-sm md:text-title-md lg:text-title-lg">
				<span className="">{t("hero.title")}</span>
				<span className="text-accent-brand">{t("hero.delivered")}</span>
			</p>
			<p className="text-muted-foreground mb-10">
				{t("hero.description")}
			</p>
			<div className="flex gap-5">
				<div className="w-32 h-12 bg-accent-brand hover:scale-90 text-white rounded-md">
					<Button text={t("hero.shopNow")} />
				</div>
				<div className="bg-secondary text-secondary-foreground hover:scale-90 rounded-md w-32 h-12">
					<Button text={t("hero.viewDeals")} />
				</div>
			</div>
		</div>
	);
}
