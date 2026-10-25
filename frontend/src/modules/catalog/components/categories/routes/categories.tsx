"use client";

import CategoriesLayout from "../components/layout/categoriesLayout";
import { CatalogFeedLoader } from "../../CatalogFeedLoader";

export function Categories() {
	return (
		<>
			<CatalogFeedLoader />
			<CategoriesLayout />
		</>
	);
}
