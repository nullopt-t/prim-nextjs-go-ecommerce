import { useEffect } from "react";
import { useTranslation } from "react-i18next";

export default function useDirection() {
	const { i18n } = useTranslation();

	useEffect(() => {
		const currentLanguage = i18n.resolvedLanguage;

		document.documentElement.lang = currentLanguage;
		document.documentElement.dir = currentLanguage === "ar" ? "rtl" : "ltr";
	}, [i18n.resolvedLanguage]);
}
