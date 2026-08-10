import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import LanguageDetector from "i18next-browser-languagedetector";

// English
import commonEn from "./locales/en/common.json";
import homeEn from "./locales/en/home.json";
import authEn from "./locales/en/auth.json";
import aboutEn from "./locales/en/about.json";
import cartEn from "./locales/en/cart.json";
import notFoundEn from "./locales/en/notFound.json";
import wishlistEn from "./locales/en/wishlist.json";
import settingsEn from "./locales/en/settings.json";

// Arabic
import commonAr from "./locales/ar/common.json";
import homeAr from "./locales/ar/home.json";
import authAr from "./locales/ar/auth.json";
import aboutAr from "./locales/ar/about.json";
import cartAr from "./locales/ar/cart.json";
import notFoundAr from "./locales/ar/notFound.json";
import wishlistAr from "./locales/ar/wishlist.json";
import settingsAr from "./locales/ar/settings.json";

i18n.use(LanguageDetector)
	.use(initReactI18next)
	.init({
		resources: {
			en: {
				common: commonEn,
				home: homeEn,
				auth: authEn,
				about: aboutEn,
				cart: cartEn,
				notFound: notFoundEn,
				wishlist: wishlistEn,
				settings: settingsEn,
			},
			ar: {
				common: commonAr,
				home: homeAr,
				auth: authAr,
				about: aboutAr,
				cart: cartAr,
				notFound: notFoundAr,
				wishlist: wishlistAr,
				settings: settingsAr,
			},
		},

		// Default language if detection fails
		fallbackLng: "en",

		// Namespace used when none is specified
		defaultNS: "common",

		// Register all namespaces
		ns: [
			"common",
			"home",
			"auth",
			"about",
			"cart",
			"notFound",
			"wishlist",
			"settings",
		],

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
