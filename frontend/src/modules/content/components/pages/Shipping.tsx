import { getTranslations } from "next-intl/server";
import { formatMoney } from "@/shared/lib/money";
import { getPublicSettings } from "@/shared/lib/server/publicSettings";
import { Link } from "@/shared/i18n/navigation";

export async function Shipping() {
	const t = await getTranslations("shop");
	const settings = await getPublicSettings();
	const currency = settings?.currency?.trim() ?? "";
	const days = settings?.estimatedDeliveryDays?.trim() ?? "—";
	const rate = settings ? formatMoney(settings.flatRateShipping, currency) : "—";
	const threshold = settings ? formatMoney(settings.freeShippingThreshold, currency) : "—";

	return (
		<div className="mx-auto flex w-full max-w-3xl flex-col gap-6 py-4 sm:py-6">
			<h1 className="text-3xl font-bold text-foreground md:text-4xl">
				{t("pages.shipping.title")}
			</h1>
			<div className="flex flex-col gap-4 text-base leading-relaxed text-muted-foreground">
				<p>{t("pages.shipping.intro")}</p>
				<h2 className="mt-4 text-xl font-semibold text-foreground">{t("pages.shipping.standardTitle")}</h2>
				<p>{t("pages.shipping.standardDescription", { days, rate, threshold })}</p>
				<h2 className="mt-4 text-xl font-semibold text-foreground">{t("pages.shipping.expressTitle")}</h2>
				<p>{t("pages.shipping.expressDescription")}</p>
				<p>
					<Link href="/contact" className="text-accent-brand hover:underline">
						{t("pages.shipping.contactCta")}
					</Link>
				</p>
			</div>
		</div>
	);
}
