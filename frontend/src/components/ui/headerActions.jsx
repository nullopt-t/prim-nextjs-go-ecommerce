import { Heart, ShoppingCart, User, Moon } from "lucide-react";
import { MdOutlineWbSunny } from "react-icons/md";
import { useTheme } from "@/context/theme";
import { useTranslation } from "react-i18next";

export function HeaderActions() {
	const { theme, toggle } = useTheme();
	const { i18n } = useTranslation();

	const currentLanguage = i18n.resolvedLanguage;

	const toggleLang = () =>
		i18n.changeLanguage(currentLanguage === "en" ? "ar" : "en");

	const icons = [
		{
			id: 1,
			icon: User,
		},
		{
			id: 2,
			icon: Heart,
		},
		{
			id: 3,
			icon: ShoppingCart,
		},
	];

	return (
		<>
			<p
				onClick={toggleLang}
				className="cursor-pointer font-medium hover:text-accent-brand"
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
			{theme === "light" ? (
				<MdOutlineWbSunny
					onClick={toggle}
					className="size-6 cursor-pointer hover:text-accent-brand"
				/>
			) : (
				<Moon
					onClick={toggle}
					className="size-6 cursor-pointer hover:text-accent-brand"
				/>
			)}
			{icons.map((value) => (
				<value.icon
					key={value.id}
					className="size-6 cursor-pointer hover:text-accent-brand"
				/>
			))}
		</>
	);
}
