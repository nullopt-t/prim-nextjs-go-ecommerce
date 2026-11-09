import type { ProductDetailsEntity } from "@/modules/catalog/types";
import type { CardDetails } from "@/shared/ui/productsCard";
import { PLACEHOLDER_PRODUCT_IMAGE } from "@/shared/lib/media";

/** Maps a catalog entity onto the card shape rendered by ProductsCard. */
export function toCardDetails(p: ProductDetailsEntity): CardDetails {
	const product =
		typeof p.product === "object"
			? { en: p.product.en ?? "", ar: p.product.ar ?? "" }
			: (() => {
					const name = String(p.product ?? p.title ?? "");
					return { en: name, ar: name };
				})();
	return {
		id: p.id,
		slug: p.slug,
		img: p.img ?? PLACEHOLDER_PRODUCT_IMAGE,
		product,
		stars: p.stars ?? 0,
		reviews: p.reviews ?? 0,
		price: p.extractedPrice ?? p.price,
		oldPrice: p.extractedOriginalPrice ?? p.oldPrice,
		discountPercentage: p.discountPercentage,
		variants: p.variants,
		brand: typeof p.brand === "string" ? p.brand : p.brand?.name,
		inStock: p.inStock,
	};
}
