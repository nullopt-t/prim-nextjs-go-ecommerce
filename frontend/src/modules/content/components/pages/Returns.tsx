import { getTranslations } from "next-intl/server";
import { Link } from "@/shared/i18n/navigation";

export async function Returns() {
	const t = await getTranslations("shop");

	return (
		<div className="mx-auto flex w-full max-w-3xl flex-col gap-6 py-4 sm:py-6">
			<h1 className="text-3xl font-bold text-foreground md:text-4xl">
				{t("pages.returns.title")}
			</h1>
			<div className="flex flex-col gap-4 text-base leading-relaxed text-muted-foreground">
				<p>{t("pages.returns.intro")}</p>
				<h2 className="mt-4 text-xl font-semibold text-foreground">{t("pages.returns.policyTitle")}</h2>
				<p>{t("pages.returns.policyDescription")}</p>
				<h2 className="mt-4 text-xl font-semibold text-foreground">{t("pages.returns.refundsTitle")}</h2>
				<p>{t("pages.returns.refundsDescription")}</p>
				<p>
					<Link href="/contact" className="text-accent-brand hover:underline">
						{t("pages.returns.contactCta")}
					</Link>
				</p>
			</div>
		</div>
	);
}
