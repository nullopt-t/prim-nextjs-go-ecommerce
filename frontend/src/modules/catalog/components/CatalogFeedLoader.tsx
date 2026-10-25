"use client";

import { useEffect, useRef } from "react";
import { useCatalogContext } from "../model/catalogContext";

/**
 * Some storefront routes skip the server-sent product feed for payload size.
 * Layouts persist across client navigation, so a feed route entered from one of
 * those pages would otherwise see an empty catalog. Fetch once when mounted
 * with no products and let the normal provider refresh path fill it in.
 */
export function CatalogFeedLoader() {
	const { products, loading, refreshCatalog } = useCatalogContext();
	const attempted = useRef(false);

	useEffect(() => {
		if (!loading && products.length === 0 && !attempted.current) {
			attempted.current = true;
			void refreshCatalog();
		}
	}, [loading, products.length, refreshCatalog]);

	return null;
}
