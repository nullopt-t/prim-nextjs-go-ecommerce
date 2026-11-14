"use client";

import CategoriesLayout from "../components/layout/categoriesLayout";
import type { Category, ProductDetailsEntity } from "@/modules/catalog/types";

export function Categories({
	categories,
	products,
}: {
	categories?: Category[];
	products?: ProductDetailsEntity[];
}) {
	return <CategoriesLayout categories={categories} products={products} />;
}
