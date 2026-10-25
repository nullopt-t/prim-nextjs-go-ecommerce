"use client";

import { useTranslations } from "next-intl";
import { usePublicSettings } from "@/modules/settings/model/publicSettingsContext";

export function Contact() {
	const t = useTranslations("shop");
	const { settings } = usePublicSettings();

	return (
		<div className="w-full flex flex-col gap-6 py-4 sm:py-6">
			<h1 className="text-3xl md:text-4xl font-bold text-foreground">
				{t("pages.contact.title")}
			</h1>
			<div className="text-muted-foreground flex flex-col gap-4 text-base leading-relaxed">
				<p>{t("pages.contact.intro")}</p>
				<div className="mt-4 p-5 bg-secondary/30 rounded-2xl border border-border space-y-2">
					<p className="font-semibold text-foreground">
						{t("pages.contact.supportLabel")}
					</p>
					{settings?.storeEmail ? (
						<a
							href={`mailto:${settings.storeEmail}`}
							className="block text-accent-brand hover:underline"
						>
							{settings.storeEmail}
						</a>
					) : null}
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
				</div>
			</div>
		</div>
	);
}
