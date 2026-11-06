import { getTranslations } from "next-intl/server";
import { getPublicSettings } from "@/shared/lib/server/publicSettings";

export async function Press() {
	const t = await getTranslations("shop");
	const settings = await getPublicSettings();

	return (
		<div className="mx-auto flex w-full max-w-3xl flex-col gap-6 py-4 sm:py-6">
			<h1 className="text-3xl font-bold text-foreground md:text-4xl">{t("pages.press.title")}</h1>
			<div className="flex flex-col gap-4 text-base leading-relaxed text-muted-foreground">
				<p>{t("pages.press.description")}</p>
				<p>
					{t("pages.press.inquiries")}{" "}
					{settings?.storeEmail ? (
						<a href={`mailto:${settings.storeEmail}`} className="text-accent-brand hover:underline">
							{settings.storeEmail}
						</a>
					) : null}
				</p>
			</div>
		</div>
	);
}
