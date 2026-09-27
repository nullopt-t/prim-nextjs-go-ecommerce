"use client";

import { Heart, ShoppingCart, User, Moon } from "lucide-react";
import { MdOutlineWbSunny } from "react-icons/md";
import { useTheme } from "@/context/theme";
import { useLocale } from "next-intl";
import { useRouter, usePathname } from "@/i18n/navigation";
import { CartSidebar } from "./cartSidebar";
import { useCart } from "@/hooks/useCart";

export function HeaderActions() {
	const { theme, toggle } = useTheme();
	const locale = useLocale();
	const router = useRouter();
	const pathname = usePathname();
	const { cartItems, openCart } = useCart();

	const currentLanguage = locale;
	const toggleLang = () => {
		const nextLocale = currentLanguage === "en" ? "ar" : "en";
		router.replace(pathname, { locale: nextLocale });
	};

	const cartCount = cartItems?.length || 0;
	// TODO: hook up real wishlist count
	const wishlistCount = 0;

	return (
		<>
			<div className="flex items-center gap-1 sm:gap-2">
				{/* Language Switcher */}
				<button
					type="button"
					onClick={toggleLang}
					dir="ltr"
					aria-label="Switch Language"
					className="h-9 px-2 sm:px-2.5 rounded-lg font-medium text-xs sm:text-sm hover:text-accent-brand hover:bg-secondary/60 transition-colors text-muted-foreground flex items-center gap-1.5 cursor-pointer"
				>
					<span className={currentLanguage === "en" ? "text-accent-brand font-bold" : "text-foreground"}>EN</span>
					<span className="text-border">|</span>
					<span className={currentLanguage === "ar" ? "text-accent-brand font-bold text-sm" : "text-foreground text-sm"}>ع</span>
				</button>

				{/* Theme Toggle */}
				<button
					type="button"
					onClick={toggle}
					aria-label="Toggle theme"
					className="size-9 rounded-lg text-muted-foreground hover:text-accent-brand hover:bg-secondary/60 transition-colors flex items-center justify-center cursor-pointer"
				>
					{theme === "light" ? <MdOutlineWbSunny className="size-4.5" /> : <Moon className="size-4.5" />}
				</button>
						
				<div className="h-4 w-px bg-border/80 mx-0.5 sm:mx-1 hidden xs:block"></div>
				
				<div className="flex items-center gap-0.5 sm:gap-1">
					{/* Account */}
					<button
						type="button"
						onClick={() => router.push("/auth")}
						aria-label="Account"
						className="size-9 rounded-lg text-muted-foreground hover:text-accent-brand hover:bg-secondary/60 transition-colors relative flex items-center justify-center cursor-pointer"
					>
						<User className="size-4.5" />
					</button>
					
					{/* Wishlist */}
					<button
						type="button"
						onClick={() => router.push("/user/wishlist")}
						aria-label="Wishlist"
						className="size-9 rounded-lg text-muted-foreground hover:text-accent-brand hover:bg-secondary/60 transition-colors relative flex items-center justify-center cursor-pointer"
					>
						<Heart className="size-4.5" />
						{wishlistCount > 0 && (
							<span className="absolute -top-0.5 ltr:-right-0.5 rtl:-left-0.5 min-w-[16px] h-4 px-1 bg-accent-brand text-white text-[10px] font-bold rounded-full flex items-center justify-center shadow-xs">
								{wishlistCount}
							</span>
						)}
					</button>
					
					{/* Cart */}
					<button
						type="button"
						onClick={openCart}
						aria-label="Shopping Cart"
						className="size-9 rounded-lg text-muted-foreground hover:text-accent-brand hover:bg-secondary/60 transition-colors relative flex items-center justify-center cursor-pointer"
					>
						<ShoppingCart className="size-4.5" />
						{cartCount > 0 && (
							<span className="absolute -top-0.5 ltr:-right-0.5 rtl:-left-0.5 min-w-[16px] h-4 px-1 bg-accent-brand text-white text-[10px] font-bold rounded-full flex items-center justify-center shadow-xs">
								{cartCount}
							</span>
						)}
					</button>
				</div>
			</div>
			
			<CartSidebar />
		</>
	);
}
