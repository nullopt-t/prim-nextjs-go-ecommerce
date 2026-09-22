import { useTranslations } from "next-intl";
import { Link } from "@/i18n/navigation";
import { Title } from "@/components/ui/title";

export default function NotFound() {
  const t = useTranslations("notFound");

  return (
    <div className="p-5 md:p-10 flex flex-col items-center justify-center min-h-[60vh]">
      <Title title={t("title")} />
      <p className="text-center text-foreground text-txt-sm md:text-txt-md lg:text-txt-lg mb-8">
        {t("subtitle")}
      </p>
      <Link
        href="/"
        className="px-6 py-3 bg-accent-brand text-white rounded-md font-medium hover:opacity-90"
      >
        {t("button")}
      </Link>
    </div>
  );
}
