import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";

export default function NotFoundActions() {
	const { t } = useTranslation("notFound");

	return (
		<div className="mt-5 flex gap-2.5 justify-center w-full h-12">
			<Link
				to="/"
				className="p-2.5 font-medium rounded-md bg-primary text-primary-foreground hover:scale-90"
			>
				{t("home")}
			</Link>
			<Link
				to="/"
				className="p-2.5 font-medium rounded-md bg-primary text-primary-foreground hover:scale-90"
			>
				{t("products")}
			</Link>
		</div>
	);
}
