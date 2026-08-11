import { Text } from "@/components/ui";
import { Toggle } from "@/components/ui/Toggle";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/context/theme";

export default function Preferences() {
	const { theme, toggle } = useTheme();
	const [emailNotification, setEmailNotification] = useState(false);

	const { i18n } = useTranslation();
	const { t } = useTranslation("settings");

	const currentLanguage = i18n.resolvedLanguage;
	const toggleLang = () =>
		i18n.changeLanguage(currentLanguage === "en" ? "ar" : "en");

	return (
		<div className="flex flex-col gap-5">
			<div className="flex justify-between">
				<Text text={t("settings.language")} />
				<p
					onClick={toggleLang}
					className="cursor-pointer font-medium text-foreground hover:text-accent-brand"
				>
					<span
						className={`${currentLanguage === "en" ? `text-accent-brand` : ``}`}
					>
						En
					</span>
					<span className=""> / </span>
					<span
						className={`${currentLanguage === "ar" ? `text-accent-brand` : ``}`}
					>
						ع
					</span>
				</p>
			</div>
			<div className="flex justify-between items-center">
				<Text text={t("settings.darkMode")} />
				<Toggle
					isEnabled={theme === "dark" ? true : false}
					onChange={toggle}
				/>
			</div>
			<div className="flex justify-between">
				<Text text={t("settings.emailNotifications")} />
				<Toggle
					isEnabled={emailNotification}
					onChange={(e) => setEmailNotification(e.target.checked)}
				/>
			</div>
		</div>
	);
}
