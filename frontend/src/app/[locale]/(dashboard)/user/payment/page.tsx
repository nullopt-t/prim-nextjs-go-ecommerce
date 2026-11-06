import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";

export async function generateMetadata(): Promise<Metadata> {
	const t = await getTranslations("payment");
	return { title: t("title"), description: t("subtitle") };
}

export default async function PaymentPage() {
	const t = await getTranslations("payment");

	return (
		<div className="flex flex-col gap-6">
			<div>
				<h1 className="text-xl font-semibold text-foreground">{t("title")}</h1>
				<p className="text-sm text-muted-foreground mt-1">{t("subtitle")}</p>
			</div>

			<div className="bg-card border border-border rounded-lg p-12 flex flex-col items-center justify-center text-center gap-2 shadow-sm">
				<p className="font-medium text-foreground">{t("comingSoonTitle")}</p>
				<p className="text-sm text-muted-foreground max-w-sm">{t("comingSoonDesc")}</p>
			</div>
		</div>
	);
}
