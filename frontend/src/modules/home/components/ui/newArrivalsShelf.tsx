"use client";

import { useMemo } from "react";
import { useTranslations } from "next-intl";
import ProductShelf from "./productShelf";
import type { ProductDetailsEntity } from "@/modules/catalog/types";

function createdAtTime(p: ProductDetailsEntity & { createdAt?: string }): number {
	const time = p.createdAt ? Date.parse(p.createdAt) : NaN;
	return Number.isNaN(time) ? 0 : time;
}

export default function NewArrivalsShelf({ products = [] }: { products?: ProductDetailsEntity[] }) {
	const t = useTranslations("home.newArrivals");

	const arrivals = useMemo(
		() => [...products].sort((a, b) => createdAtTime(b) - createdAtTime(a)).slice(0, 8),
		[products]
	);

	if (arrivals.length === 0) return null;

	return (
		<ProductShelf
			title={t("title")}
			subtitle={t("subtitle")}
			link="/products"
			products={arrivals}
		/>
	);
}
