import { useTranslation } from "react-i18next";

export default function BeliefsCards({ belief }) {
	const { t } = useTranslation("about");

	return (
		<div className="p-2.5 md:p-5 bg-card rounded-lg">
			<belief.icon className="mb-2.5 text-accent-brand"></belief.icon>
			<p className="mb-2.5 text-card-foreground font-medium text-txt-sm md:text-txt-md lg:text-txt-lg">
				{t(belief.title)}
			</p>
			<p className="text-muted-foreground">{t(belief.subTitle)}</p>
		</div>
	);
}
