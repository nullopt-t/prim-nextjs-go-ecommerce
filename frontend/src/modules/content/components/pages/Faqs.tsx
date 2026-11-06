import { getTranslations } from "next-intl/server";
import { Link } from "@/shared/i18n/navigation";

export async function Faqs() {
	const t = await getTranslations("shop");

	const faqs = [
		{ question: t("pages.faqs.trackOrder.question"), answer: t("pages.faqs.trackOrder.answer") },
		{ question: t("pages.faqs.changeOrder.question"), answer: t("pages.faqs.changeOrder.answer") },
		{ question: t("pages.faqs.payment.question"), answer: t("pages.faqs.payment.answer") },
	];

	const jsonLd = {
		"@context": "https://schema.org",
		"@type": "FAQPage",
		mainEntity: faqs.map((faq) => ({
			"@type": "Question",
			name: faq.question,
			acceptedAnswer: { "@type": "Answer", text: faq.answer },
		})),
	};

	return (
		<div className="mx-auto flex w-full max-w-3xl flex-col gap-6 py-4 sm:py-6">
			<script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }} />
			<h1 className="text-3xl font-bold text-foreground md:text-4xl">{t("pages.faqs.title")}</h1>
			<div className="flex flex-col divide-y divide-border rounded-lg border border-border">
				{faqs.map((faq) => (
					<details key={faq.question} className="group px-5 py-4">
						<summary className="cursor-pointer text-base font-semibold text-foreground marker:text-muted-foreground">
							{faq.question}
						</summary>
						<p className="mt-2 text-base leading-relaxed text-muted-foreground">{faq.answer}</p>
					</details>
				))}
			</div>
			<p className="text-sm text-muted-foreground">
				<Link href="/contact" className="text-accent-brand hover:underline">
					{t("pages.faqs.contactCta")}
				</Link>
			</p>
		</div>
	);
}
