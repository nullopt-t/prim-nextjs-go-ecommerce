import { NavLink } from "react-router-dom";
import { useTranslation } from "react-i18next";

export function Categories() {
	const { t, i18n } = useTranslation("common");

	const categories = [
		{
			id: "cat_1",
			slug: "audio",
			path: "/product",
			name: {
				en: "Audio",
				ar: "الأجهزة الصوتية",
			},
		},
		{
			id: "cat_2",
			slug: "wearables",
			path: "/product",
			name: {
				en: "Wearables",
				ar: "الأجهزة القابلة للارتداء",
			},
		},
		{
			id: "cat_3",
			slug: "desk-setup",
			path: "/product",
			name: {
				en: "Desk Setup",
				ar: "مستلزمات المكتب",
			},
		},
		{
			id: "cat_4",
			slug: "accessories",
			path: "/product",
			name: {
				en: "Accessories",
				ar: "الإكسسوارات",
			},
		},
		{
			id: "cat_5",
			slug: "lighting",
			path: "/product",
			name: {
				en: "Lighting",
				ar: "وحدات الإضاءة",
			},
		},
	];

	const constCategories = [
		{
			id: "home-232f",
			slug: "Home",
			path: "/home",
			label: "header.home",
		},
		{
			id: "all-catg-2389",
			slug: "All categories",
			path: "/all-categories",
			label: "header.allCategories",
		},
	];

	return (
		<ul className="flex gap-2.5 md:gap-5 overflow-auto text-txt-sm md:text-txt-md lg:text-txt-lg">
			{constCategories.map((value) => (
				<li key={value.id} className="cursor-pointer group">
					<NavLink
						to={value.path}
						className={({
							isActive,
						}) => `group-hover:text-accent-brand
						${isActive ? `text-accent-brand underline underline-offset-8 decoration-2 decoration-accent-brand` : `text-foreground`}`}
					>
						{t(value.label)}
					</NavLink>
				</li>
			))}
			{categories.map((value) => (
				<li key={value.id} className="cursor-pointer group">
					<NavLink
						to="/product"
						className={({
							isActive,
						}) => `group-hover:text-accent-brand
						${isActive ? `text-accent-brand underline underline-offset-8 decoration-2 decoration-accent-brand` : `text-foreground`}`}
					>
						{i18n.resolvedLanguage === "en"
							? value.name.en
							: value.name.ar}
					</NavLink>
				</li>
			))}
		</ul>
	);
}
