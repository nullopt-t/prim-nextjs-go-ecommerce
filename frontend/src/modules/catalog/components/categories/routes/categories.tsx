"use client";

import CategoriesLayout from "../components/layout/categoriesLayout";

export function Categories({ categories }: { categories?: import("@/modules/catalog/types").Category[] }) {
	return <CategoriesLayout categories={categories} />;
}
