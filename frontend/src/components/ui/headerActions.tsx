"use client";

import { Heart, ShoppingCart, User, Moon } from "lucide-react";
import { MdOutlineWbSunny } from "react-icons/md";
import { useTheme } from "@/context/theme";
import { useLocale } from "next-intl";
import { useRouter, usePathname, Link } from "@/i18n/navigation";
import { useCart } from "@/hooks/useCart";

export function HeaderActions() {
  const { theme, toggle } = useTheme();
  const locale = useLocale();
  const router = useRouter();
  const pathname = usePathname();
  const { cartItems, toggleCart } = useCart();
  const cartCount = cartItems?.reduce((sum, item) => sum + (item.quantity || 1), 0) || 0;

  const toggleLang = () => {
    const nextLocale = locale === "en" ? "ar" : "en";
    router.replace(pathname, { locale: nextLocale });
  };

  return (
    <div className="flex items-center gap-3 sm:gap-4 md:gap-5">
      <button
        type="button"
        onClick={toggleLang}
        className="cursor-pointer text-sm font-semibold hover:text-accent-brand px-1 py-1"
      >
        <span className={locale === "en" ? "text-accent-brand font-bold" : ""}>
          En
        </span>
        <span> / </span>
        <span className={locale === "ar" ? "text-accent-brand font-bold" : ""}>
          ع
        </span>
      </button>

      <button
        type="button"
        onClick={toggle}
        aria-label="Toggle theme"
        className="cursor-pointer hover:text-accent-brand"
      >
        {theme === "light" ? (
          <MdOutlineWbSunny className="size-6" />
        ) : (
          <Moon className="size-6" />
        )}
      </button>

      <Link href="/auth" aria-label="Account" className="hover:text-accent-brand">
        <User className="size-6" />
      </Link>

      <Link
        href="/user/wishlist"
        aria-label="Wishlist"
        className="hover:text-accent-brand"
      >
        <Heart className="size-6" />
      </Link>

      <button
        type="button"
        onClick={toggleCart}
        aria-label="Cart"
        className="relative hover:text-accent-brand cursor-pointer"
      >
        <ShoppingCart className="size-6" />
        {cartCount > 0 && (
          <span className="absolute -top-1.5 -right-2 bg-accent-brand text-white font-bold text-[10px] rounded-full h-4 min-w-[16px] px-1 flex items-center justify-center shadow-xs">
            {cartCount > 99 ? "99+" : cartCount}
          </span>
        )}
      </button>
    </div>
  );
}
