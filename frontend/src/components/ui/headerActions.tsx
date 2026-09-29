"use client";

import { useState, useRef, useEffect } from "react";
import {
	ShoppingCart,
	User,
	Heart,
	Package,
	Settings,
	LogOut,
	LogIn,
	Languages,
	Sun,
	Moon,
	ChevronDown,
} from "lucide-react";
import { useTheme } from "@/context/theme";
import { useLocale, useTranslations } from "next-intl";
import { useRouter, usePathname, Link } from "@/i18n/navigation";
import { useCart } from "@/hooks/useCart";
import { useAuthContext } from "@/context/AuthContext";

export function HeaderActions() {
	const { theme, toggle } = useTheme();
	const locale = useLocale();
	const router = useRouter();
	const pathname = usePathname();
	const { cartItems, openCart } = useCart();
	const { user, isAuthenticated, logout } = useAuthContext();
	const t = useTranslations("common.header");

	const [isMenuOpen, setIsMenuOpen] = useState(false);
	const dropdownRef = useRef<HTMLDivElement>(null);

	const cartCount = cartItems?.length || 0;

	// Close dropdown when clicking outside or pressing Escape
	useEffect(() => {
		const handleClickOutside = (event: MouseEvent) => {
			if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
				setIsMenuOpen(false);
			}
		};
		const handleKeyDown = (event: KeyboardEvent) => {
			if (event.key === "Escape") {
				setIsMenuOpen(false);
			}
		};

		if (isMenuOpen) {
			document.addEventListener("mousedown", handleClickOutside);
			document.addEventListener("keydown", handleKeyDown);
		}
		return () => {
			document.removeEventListener("mousedown", handleClickOutside);
			document.removeEventListener("keydown", handleKeyDown);
		};
	}, [isMenuOpen]);

	const toggleLanguage = () => {
		const nextLocale = locale === "en" ? "ar" : "en";
		router.replace(pathname, { locale: nextLocale });
	};

	const handleSignOut = async () => {
		setIsMenuOpen(false);
		await logout();
		router.push("/");
	};

	return (
		<div className="flex items-center gap-2 sm:gap-2.5">
			{/* 1. Shopping Cart Button */}
			<button
				type="button"
				onClick={openCart}
				aria-label={t("cart")}
				className="size-9 sm:size-10 rounded-xl text-muted-foreground hover:text-accent-brand hover:bg-secondary/70 transition-all relative flex items-center justify-center cursor-pointer border border-border/40 hover:border-accent-brand/40 shadow-2xs"
			>
				<ShoppingCart className="size-4.5 sm:size-5" />
				{cartCount > 0 && (
					<span className="absolute -top-1 ltr:-right-1 rtl:-left-1 min-w-[18px] h-4.5 px-1 bg-accent-brand text-white text-[10px] font-bold rounded-full flex items-center justify-center shadow-xs ring-2 ring-background animate-in zoom-in duration-150">
						{cartCount}
					</span>
				)}
			</button>

			{/* 2. User Menu & Settings Dropdown */}
			<div className="relative" ref={dropdownRef}>
				<button
					type="button"
					onClick={() => setIsMenuOpen((prev) => !prev)}
					aria-expanded={isMenuOpen}
					aria-haspopup="true"
					aria-label={t("account")}
					className={`h-9 sm:h-10 px-2 sm:px-2.5 rounded-xl transition-all flex items-center gap-1.5 cursor-pointer border text-sm font-medium ${
						isMenuOpen
							? "border-accent-brand/60 bg-secondary/80 text-foreground shadow-xs"
							: "border-border/40 text-muted-foreground hover:text-accent-brand hover:bg-secondary/70 hover:border-accent-brand/40 shadow-2xs"
					}`}
				>
					<div className="size-6 sm:size-6.5 rounded-full bg-accent-brand/10 text-accent-brand flex items-center justify-center font-semibold text-xs">
						{isAuthenticated && user?.name ? (
							user.name.charAt(0).toUpperCase()
						) : (
							<User className="size-3.5 sm:size-4" />
						)}
					</div>
					{isAuthenticated && user?.name && (
						<span className="hidden md:inline-block max-w-[80px] truncate text-foreground font-medium text-xs">
							{user.name.split(" ")[0]}
						</span>
					)}
					<ChevronDown
						className={`size-3.5 text-muted-foreground transition-transform duration-200 ${
							isMenuOpen ? "rotate-180 text-accent-brand" : ""
						}`}
					/>
				</button>

				{/* Dropdown Menu Panel */}
				{isMenuOpen && (
					<div
						className={`absolute mt-2 w-64 rounded-2xl bg-card border border-border/80 shadow-xl py-2 z-50 animate-in fade-in-0 zoom-in-95 duration-150 ${
							locale === "ar" ? "left-0 origin-top-left" : "right-0 origin-top-right"
						}`}
					>
						{/* User Info Header or Guest Header */}
						<div className="px-4 py-2.5 border-b border-border/60">
							{isAuthenticated && user ? (
								<div>
									<p className="font-semibold text-foreground text-sm truncate">
										{user.name || "Customer"}
									</p>
									<p className="text-muted-foreground text-xs truncate">{user.email}</p>
								</div>
							) : (
								<div className="flex items-center justify-between">
									<div>
										<p className="font-medium text-foreground text-sm">{t("account")}</p>
										<p className="text-muted-foreground text-xs">Welcome to PRIM</p>
									</div>
									<Link
										href="/auth"
										onClick={() => setIsMenuOpen(false)}
										className="inline-flex items-center gap-1 text-xs font-semibold px-2.5 py-1.5 rounded-lg bg-accent-brand text-white hover:bg-accent-brand/90 transition-colors shadow-2xs"
									>
										<LogIn className="size-3.5" />
										{t("signIn")}
									</Link>
								</div>
							)}
						</div>

						{/* Primary Navigation Links */}
						<div className="py-1 px-1.5 text-xs font-medium space-y-0.5">
							<Link
								href="/wishlist"
								onClick={() => setIsMenuOpen(false)}
								className="flex items-center gap-2.5 px-3 py-2 rounded-xl text-foreground hover:text-accent-brand hover:bg-secondary/70 transition-colors"
							>
								<Heart className="size-4 text-muted-foreground" />
								<span>{t("wishlist")}</span>
							</Link>

							{isAuthenticated && (
								<>
									<Link
										href="/user/orders"
										onClick={() => setIsMenuOpen(false)}
										className="flex items-center gap-2.5 px-3 py-2 rounded-xl text-foreground hover:text-accent-brand hover:bg-secondary/70 transition-colors"
									>
										<Package className="size-4 text-muted-foreground" />
										<span>{t("orders")}</span>
									</Link>

									<Link
										href="/user/settings"
										onClick={() => setIsMenuOpen(false)}
										className="flex items-center gap-2.5 px-3 py-2 rounded-xl text-foreground hover:text-accent-brand hover:bg-secondary/70 transition-colors"
									>
										<Settings className="size-4 text-muted-foreground" />
										<span>{t("settings")}</span>
									</Link>
								</>
							)}
						</div>

						{/* Preferences Section: Language & Theme */}
						<div className="my-1 border-t border-border/60 pt-1.5 px-1.5 space-y-1">
							{/* Language Switch */}
							<button
								type="button"
								onClick={toggleLanguage}
								className="w-full flex items-center justify-between px-3 py-2 rounded-xl text-xs font-medium text-foreground hover:bg-secondary/70 transition-colors cursor-pointer"
							>
								<div className="flex items-center gap-2.5">
									<Languages className="size-4 text-muted-foreground" />
									<span>{t("language")}</span>
								</div>
								<div className="flex items-center gap-1.5 text-[11px] font-semibold bg-secondary/80 px-2 py-0.5 rounded-md border border-border/50">
									<span className={locale === "en" ? "text-accent-brand font-bold" : "text-muted-foreground"}>
										EN
									</span>
									<span className="text-border">|</span>
									<span className={locale === "ar" ? "text-accent-brand font-bold" : "text-muted-foreground"}>
										عربي
									</span>
								</div>
							</button>

							{/* Theme Switch */}
							<button
								type="button"
								onClick={toggle}
								className="w-full flex items-center justify-between px-3 py-2 rounded-xl text-xs font-medium text-foreground hover:bg-secondary/70 transition-colors cursor-pointer"
							>
								<div className="flex items-center gap-2.5">
									{theme === "light" ? (
										<Sun className="size-4 text-amber-500" />
									) : (
										<Moon className="size-4 text-indigo-400" />
									)}
									<span>{t("theme")}</span>
								</div>
								<span className="text-[11px] text-muted-foreground capitalize font-medium bg-secondary/80 px-2 py-0.5 rounded-md border border-border/50">
									{theme === "light" ? t("light") : t("dark")}
								</span>
							</button>
						</div>

						{/* Sign out if authenticated */}
						{isAuthenticated && (
							<div className="border-t border-border/60 pt-1 px-1.5">
								<button
									type="button"
									onClick={handleSignOut}
									className="w-full flex items-center gap-2.5 px-3 py-2 rounded-xl text-xs font-medium text-destructive hover:bg-destructive/10 transition-colors cursor-pointer"
								>
									<LogOut className="size-4 text-destructive" />
									<span>{t("signOut")}</span>
								</button>
							</div>
						)}
					</div>
				)}
			</div>
		</div>
	);
}

