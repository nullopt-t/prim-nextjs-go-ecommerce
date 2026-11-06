import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import HomeLayout from "@/modules/home/components/layout/homeLayout";
import { getCatalogSnapshot } from "@/shared/lib/server/catalog";
import { getPublicSettings } from "@/shared/lib/server/publicSettings";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("meta.home");
  return { title: t("title"), description: t("description") };
}

export default async function HomePage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  const settings = await getPublicSettings();
  const storeName = settings?.storeName?.trim() || undefined;
  const snapshot = await getCatalogSnapshot(24).catch(() => ({ products: [], categories: [], brands: [] }));

  const jsonLd = {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "WebSite",
        name: storeName,
        url: `/${locale}`,
        potentialAction: {
          "@type": "SearchAction",
          target: `/${locale}/products?q={search_term_string}`,
          "query-input": "required name=search_term_string",
        },
      },
      ...(storeName
        ? [{ "@type": "Organization", name: storeName, url: `/${locale}` }]
        : []),
    ],
  };

  return (
    <>
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }} />
      <HomeLayout products={snapshot.products} categories={snapshot.categories} />
    </>
  );
}
