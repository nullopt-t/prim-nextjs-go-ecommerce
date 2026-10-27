import React from "react";
import { headers } from "next/headers";
import MainLayout from "@/shared/layouts/mainLayout";
import { MaintenanceGate } from "@/shared/ui";
import Recently from "@/modules/home/components/ui/recently";
import { CatalogProvider } from "@/modules/catalog/model/catalogContext";
import {
	getCatalogFacets,
	getCatalogSnapshot,
} from "@/shared/lib/server/catalog";
import { RouteMessages } from "@/shared/i18n/messages";

function needsProductFeed(pathname: string) {
	const path =
		pathname.split("?")[0].replace(/^\/(en|ar)(?=\/|$)/, "") || "/";
	return path === "/" || /^\/products\/[^/]+\/?$/.test(path);
}

function feedPageSize(pathname: string): number {
	const path =
		pathname.split("?")[0].replace(/^\/(en|ar)(?=\/|$)/, "") || "/";
	// Product detail only needs enough products for the "similar products"
	// carousel (5 items). Home needs shelves + the recently-viewed strip.
	return 24;
}

export default async function ShopLayout({
	children,
}: {
	children: React.ReactNode;
}) {
	const pathname = (await headers()).get("x-pathname") || "/";
	const { products, categories, brands } = needsProductFeed(pathname)
		? await getCatalogSnapshot(feedPageSize(pathname))
		: { products: [], ...(await getCatalogFacets()) };

	return (
		<RouteMessages group="shop">
			<CatalogProvider
				initialProducts={products}
				initialCategories={categories}
				initialBrands={brands}
			>
				<MaintenanceGate>
					<MainLayout recently={<Recently />}>{children}</MainLayout>
				</MaintenanceGate>
			</CatalogProvider>
		</RouteMessages>
	);
}
