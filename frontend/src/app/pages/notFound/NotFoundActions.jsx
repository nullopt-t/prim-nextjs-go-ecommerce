import { useTranslation } from "react-i18next";
import { CustomButton } from "@/components/ui";
import { useNavigate } from "react-router-dom";

export default function NotFoundActions() {
	const navigate = useNavigate();
	const { t } = useTranslation("notFound");

	return (
		<div className="flex gap-2.5 justify-center w-full h-12">
			<div className="p-2.5 rounded-md bg-primary text-primary-foreground hover:scale-90">
				<CustomButton
					text={t("home")}
					onClick={() => {
						navigate("/");
					}}
				/>
			</div>
			<div className="p-2.5 rounded-md bg-primary text-primary-foreground hover:scale-90">
				<CustomButton
					text={t("products")}
					onClick={() => {
						navigate("/");
					}}
				/>
			</div>
		</div>
	);
}
