import React from "react";
import MainLayout from "@/shared/layouts/mainLayout";
import { MaintenanceGate } from "@/shared/ui";
import Recently from "@/modules/home/components/ui/recently";
import { CatalogProvider } from "@/modules/catalog/model/catalogContext";
import { getCatalogSnapshot } from "@/shared/lib/server/catalog";

export default async function ShopLayout({
	children,
}: {
	children: React.ReactNode;
}) {
	const { products, categories, brands } = await getCatalogSnapshot();

	return (
		<CatalogProvider
			initialProducts={products}
			initialCategories={categories}
			initialBrands={brands}
		>
			<MaintenanceGate>
				<MainLayout recently={<Recently />}>{children}</MainLayout>
			</MaintenanceGate>
		</CatalogProvider>
	);
}
