"use client";

import { Link } from "@/i18n/navigation";
import { House, Moon } from "lucide-react";
import { MdOutlineWbSunny } from "react-icons/md";
import { useTheme } from "@/context/theme";
import { useTranslations } from "next-intl";

export default function FormTitle({ type }: { type: string }) {
  const { theme, toggle } = useTheme();
  const t = useTranslations("auth");

  const title =
    type === "login"
      ? t("sign.title")
      : type === "verify"
      ? t("verify.title")
      : "";

  return (
    <div className="flex justify-between items-center capitalize text-foreground font-medium text-title-sm md:text-title-md lg:text-title-lg">
      <span>{title}</span>

      <div className="flex items-center gap-3 text-muted-foreground">
        <Link
          href="/"
          aria-label="Home"
          className="hover:scale-105 hover:text-foreground transition-transform"
        >
          <House className="size-5" />
        </Link>
        <button
          type="button"
          onClick={toggle}
          aria-label="Toggle theme"
          className="hover:scale-105 hover:text-foreground transition-transform cursor-pointer"
        >
          {theme === "light" ? (
            <MdOutlineWbSunny className="size-5" />
          ) : (
            <Moon className="size-5" />
          )}
        </button>
      </div>
    </div>
  );
}
