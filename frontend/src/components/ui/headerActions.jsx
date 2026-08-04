import { Heart, ShoppingCart, User, Moon } from "lucide-react";
import { MdOutlineWbSunny } from "react-icons/md";
import { useTheme } from "@/context/theme";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

export function HeaderActions() {
	const { theme, toggle } = useTheme();
	const { i18n } = useTranslation();
	const navigate = useNavigate();

	const currentLanguage = i18n.resolvedLanguage;

	const toggleLang = () =>
		i18n.changeLanguage(currentLanguage === "en" ? "ar" : "en");

	const icons = [
		{
			id: 1,
			icon: User,
			path: "/auth",
		},
		{
			id: 2,
			icon: Heart,
			path: "/wishlist",
		},
		{
			id: 3,
			icon: ShoppingCart,
			path: "/cart",
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
					onClick={() => {
						navigate(value.path);
					}}
					className="size-6 cursor-pointer hover:text-accent-brand"
				/>
			))}
		</>
	);
}
