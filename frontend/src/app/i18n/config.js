import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import LanguageDetector from "i18next-browser-languagedetector";

// English
import commonEn from "./locales/en/common.json";
import homeEn from "./locales/en/home.json";

// Arabic
import commonAr from "./locales/ar/common.json";
import homeAr from "./locales/ar/home.json";

i18n.use(LanguageDetector)
	.use(initReactI18next)
	.init({
		resources: {
			en: {
				common: commonEn,
				home: homeEn,
			},
			ar: {
				common: commonAr,
				home: homeAr,
			},
		},

		// Default language if detection fails
		fallbackLng: "en",

		// Namespace used when none is specified
		defaultNS: "common",

		// Register all namespaces
		ns: ["common", "home"],

		// Language detection settings
		detection: {
			order: ["localStorage", "navigator"],
			caches: ["localStorage"],
		},

		interpolation: {
			escapeValue: false,
		},
	});

export default i18n;
