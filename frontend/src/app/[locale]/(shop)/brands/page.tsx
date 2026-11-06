import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { Link } from "@/shared/i18n/navigation";
import { getBrands } from "@/shared/lib/server/catalog";
import { normalizeMediaUrl } from "@/shared/lib/media";
import { localizedName } from "@/shared/lib/localizedName";

export async function generateMetadata(): Promise<Metadata> {
	const t = await getTranslations("meta.shop.brands");
	return { title: t("title"), description: t("description") };
}

function brandSlug(brand: { id: string; name: string }): string {
	return brand.name.toLowerCase().trim().replace(/\s+/g, "-") || brand.id;
}

export default async function BrandsPage() {
	const t = await getTranslations("shop");
	const brands = await getBrands(100);

	const jsonLd = {
		"@context": "https://schema.org",
		"@type": "ItemList",
		itemListElement: brands.map((brand, index) => ({
			"@type": "ListItem",
			position: index + 1,
			url: `/brands/${brandSlug(brand)}`,
			name: localizedName(brand.name, "en"),
		})),
	};

	return (
		<div className="w-full py-4 sm:py-6">
			<script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }} />
			<h1 className="text-3xl font-bold text-foreground md:text-4xl">{t("brands.title")}</h1>
			<p className="mt-2 text-sm text-muted-foreground">{t("brands.description")}</p>

			{brands.length === 0 ? (
				<div className="mt-8 rounded-lg border border-border bg-background px-4 py-16 text-center">
					<p className="text-sm font-medium text-foreground">{t("brands.emptyTitle")}</p>
					<p className="mt-1 text-sm text-muted-foreground">{t("brands.emptyDescription")}</p>
				</div>
			) : (
				<ul className="mt-8 grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
					{brands.map((brand) => (
						<li key={brand.id}>
							<Link
								href={`/brands/${brandSlug(brand)}`}
								className="flex items-center gap-3 rounded-lg border border-border bg-background p-4 transition-colors hover:bg-secondary/40"
							>
								{brand.logoUrl || brand.logo ? (
									<img
										src={normalizeMediaUrl(brand.logoUrl || brand.logo)}
										alt=""
										loading="lazy"
										className="size-10 shrink-0 rounded-md border border-border object-cover"
									/>
								) : (
									<span className="flex size-10 shrink-0 items-center justify-center rounded-md border border-border bg-secondary/40 text-sm font-semibold text-foreground">
										{brand.name.slice(0, 1).toUpperCase()}
									</span>
								)}
								<span className="min-w-0">
									<span dir="auto" className="block truncate text-sm font-medium text-foreground">
										{brand.name}
									</span>
									{brand.productCount !== undefined ? (
										<span className="block text-xs text-muted-foreground">
											{t("brands.productsCount", { count: brand.productCount })}
										</span>
									) : null}
								</span>
							</Link>
						</li>
					))}
				</ul>
			)}
		</div>
	);
}
