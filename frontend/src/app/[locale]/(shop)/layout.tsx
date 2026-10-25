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
	return (
		path === "/" ||
		path === "/categories" ||
		/^\/products\/[^/]+\/?$/.test(path)
	);
}

export default async function ShopLayout({
	children,
}: {
	children: React.ReactNode;
}) {
	const pathname = (await headers()).get("x-pathname") || "/";
	const { products, categories, brands } = needsProductFeed(pathname)
		? await getCatalogSnapshot()
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
