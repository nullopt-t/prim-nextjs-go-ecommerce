import { getTranslations } from "next-intl/server";
import { getPublicSettings } from "@/shared/lib/server/publicSettings";

export async function Contact() {
	const t = await getTranslations("shop");
	const settings = await getPublicSettings();

	return (
		<div className="mx-auto flex w-full max-w-3xl flex-col gap-6 py-4 sm:py-6">
			<h1 className="text-3xl font-bold text-foreground md:text-4xl">
				{t("pages.contact.title")}
			</h1>
			<div className="flex flex-col gap-4 text-base leading-relaxed text-muted-foreground">
				<p>{t("pages.contact.intro")}</p>
				<div className="mt-4 space-y-2 rounded-2xl border border-border bg-secondary/30 p-5">
					<p className="font-semibold text-foreground">{t("pages.contact.supportLabel")}</p>
					{settings?.storeEmail ? (
						<a
							href={`mailto:${settings.storeEmail}`}
							className="block text-accent-brand hover:underline"
						>
							{settings.storeEmail}
						</a>
					) : (
						<p>{t("pages.contact.email")}</p>
					)}
					{settings?.storePhone ? (
						<a
							href={`tel:${settings.storePhone}`}
							dir="ltr"
							className="block text-accent-brand hover:underline"
						>
							{settings.storePhone}
						</a>
					) : null}
					{settings?.storeAddress ? <p>{settings.storeAddress}</p> : null}
					<p>{t("pages.contact.hours")}</p>
				</div>
			</div>
		</div>
	);
}
