import { useTranslation } from "react-i18next";

export default function Title({ title }) {
	const { t } = useTranslation("about");

	return (
		<p className="mb-5 font-medium text-title-lg md:text-title-md lg:text-title-lg text-foreground">
			{t(title)}
		</p>
	);
}
