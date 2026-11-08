import { normalizeMediaUrl } from "@/shared/lib/media";
import { formatMoney } from "@/shared/lib/money";

/**
 * Shape-agnostic normalizer for catalog payloads coming from the Go API.
 * Kept free of `"use client"` so both Server and Client Components can use it.
 *
 * The raw payload can arrive in several shapes (backend product, wishlist item
 * envelope, legacy mock), so it is typed permissively with an index signature
 * rather than `any`.
 */

export interface RawProductMedia {
	url?: string;
	publicUrl?: string;
	object?: { publicUrl?: string };
	[key: string]: unknown;
}

export interface RawProductVariant {
	id?: string;
	sku?: string;
	title?: string;
	price?: string | number | null;
	extractedPrice?: number | null;
	originalPrice?: string;
	oldPrice?: string | number;
	extractedOriginalPrice?: number;
	stock?: { isInStock?: boolean; availableQuantity?: number };
	thumbnail?: string | { url?: string; publicUrl?: string };
	media?: RawProductMedia[];
	attributes?: Record<string, unknown>;
	color?: string;
	[key: string]: unknown;
}

export interface RawBundleItem {
	id?: string;
	title?: string;
	slug?: string;
	thumbnail?: string;
	quantity?: number;
	variant?: string;
	availableQuantity?: number;
	isInStock?: boolean;
}

export interface RawBundlePayload {
	itemCount?: number;
	availableQuantity?: number;
	isInStock?: boolean;
	items?: RawBundleItem[];
}

export interface RawProductPayload {
	[key: string]: unknown;
	id?: string;
	productId?: string;
	slug?: string;
	wishlistItemId?: string;
	title?: string | { en?: string; ar?: string };
	product?: { title?: string; slug?: string; en?: string; ar?: string; [key: string]: unknown };
	name?: string;
	categoryName?: string;
	brandName?: string;
	price?: string | number;
	extractedPrice?: number;
	originalPrice?: string;
	oldPrice?: string | number;
	extractedOriginalPrice?: number;
	discountPercentage?: string | number;
	rating?: { averageRating?: number; reviewCount?: number };
	stars?: number | string;
	reviews?: number | string;
	thumbnail?: string | { url?: string; publicUrl?: string };
	thumbnailUrl?: string;
	img?: string;
	image?: string;
	images?: Array<string | { url?: string; publicUrl?: string }>;
	variants?: RawProductVariant[];
	variationAxes?: string[];
	category?: string | { id?: string; name?: string };
	categories?: Array<string | { id?: string; name?: string }>;
	inStock?: boolean;
	stockCount?: number;
	currency?: string;
	variantId?: string;
	defaultVariantId?: string;
	brand?: string | { logoUrl?: string; logo?: string; [key: string]: unknown };
	bundle?: RawBundlePayload;
}

export function normalizeProduct(raw: RawProductPayload | null) {
	if (!raw) return null;

	// Handle both backend Go format, WishlistItem envelope, and legacy mock format
	const isWishlistItem = Boolean(
		raw.product && typeof raw.product === "object" && (raw.product.title || raw.product.slug)
	);
	const p: RawProductPayload = isWishlistItem ? (raw.product ?? raw) : raw;
	const wishlistItemId = isWishlistItem ? raw.id : raw.wishlistItemId || undefined;

	const id = p.id || raw.productId || p.slug || raw.id || "";
	const slug = p.slug || raw.slug || id;
	const titleStr =
		typeof p.title === "string"
			? p.title
			: typeof p.product?.en === "string"
				? p.product.en
				: typeof p.product?.ar === "string"
					? p.product.ar
					: typeof p.name === "string"
						? p.name
						: typeof raw.title === "string"
							? raw.title
							: "Product";

	const product = {
		en: typeof p.product?.en === "string" ? p.product.en : titleStr,
		ar: typeof p.product?.ar === "string" ? p.product.ar : titleStr,
	};

	const firstImage = Array.isArray(p.images) ? p.images[0] : undefined;
	const firstImageStr =
		typeof firstImage === "string" ? firstImage : firstImage?.url || firstImage?.publicUrl;

	const rawImg =
		p.thumbnailUrl ||
		(typeof p.thumbnail === "string" ? p.thumbnail : p.thumbnail?.publicUrl || p.thumbnail?.url) ||
		p.img ||
		p.image ||
		firstImageStr ||
		(typeof raw.thumbnail === "string" ? raw.thumbnail : raw.thumbnail?.publicUrl || raw.thumbnail?.url) ||
		raw.img;
	const img = normalizeMediaUrl(rawImg);

	// Price formatting - check p (inner product if wishlistItem, or raw) then fallback to raw
	const rawPrice = p.extractedPrice != null ? p.extractedPrice : raw.extractedPrice;
	const priceNum =
		rawPrice !== undefined
			? Number(rawPrice)
			: typeof p.price === "number"
				? p.price
				: typeof raw.price === "number"
					? raw.price
					: parseFloat(String(p.price || raw.price || "0").replace(/[^0-9.-]+/g, "")) || 0;

	const rawOldPrice =
		p.extractedOriginalPrice != null
			? p.extractedOriginalPrice
			: raw.extractedOriginalPrice;
	const rawOrigPriceStr = p.originalPrice || raw.originalPrice || p.oldPrice || raw.oldPrice;
	const oldPriceNum =
		rawOldPrice !== undefined
			? Number(rawOldPrice)
			: rawOrigPriceStr
				? parseFloat(String(rawOrigPriceStr).replace(/[^0-9.-]+/g, ""))
				: undefined;

	let discountPercentage = p.discountPercentage || raw.discountPercentage;
	if (!discountPercentage && oldPriceNum && oldPriceNum > priceNum) {
		discountPercentage = `${Math.round(((oldPriceNum - priceNum) / oldPriceNum) * 100)}%`;
	}

	const stars =
		(p.rating?.averageRating ?? raw.rating?.averageRating) !== undefined
			? (p.rating?.averageRating ?? raw.rating?.averageRating)
			: (p.stars ?? raw.stars) !== undefined
				? Number(p.stars ?? raw.stars)
				: 0;

	const reviews =
		(p.rating?.reviewCount ?? raw.rating?.reviewCount) !== undefined
			? (p.rating?.reviewCount ?? raw.rating?.reviewCount)
			: (p.reviews ?? raw.reviews) !== undefined
				? Number(p.reviews ?? raw.reviews)
				: 0;

	const variantMedia: string[] = [];
	if (Array.isArray(raw.variants)) {
		raw.variants.forEach((v: RawProductVariant) => {
			const vThumbRaw =
				typeof v.thumbnail === "string" ? v.thumbnail : v.thumbnail?.url || v.thumbnail?.publicUrl;
			const vThumb = normalizeMediaUrl(vThumbRaw);
			if (vThumb && vThumb !== "/placeholder-product.png" && !variantMedia.includes(vThumb)) {
				variantMedia.push(vThumb);
			}
			if (Array.isArray(v.media)) {
				v.media.forEach((m: RawProductMedia) => {
					const mUrlRaw = typeof m === "string" ? m : m?.url || m?.publicUrl || m?.object?.publicUrl;
					const mUrl = normalizeMediaUrl(mUrlRaw);
					if (mUrl && mUrl !== "/placeholder-product.png" && !variantMedia.includes(mUrl)) {
						variantMedia.push(mUrl);
					}
				});
			}
		});
	}

	// Include master product thumbnail if available
	if (img && img !== "/placeholder-product.png" && !variantMedia.includes(img)) {
		variantMedia.push(img);
	}

	const rawImagesList =
		Array.isArray(raw.images) && raw.images.length > 0
			? raw.images.map((im: string | { url?: string; publicUrl?: string }) =>
					normalizeMediaUrl(typeof im === "string" ? im : im?.url || im?.publicUrl)
				)
			: [];

	const images =
		variantMedia.length > 0 ? variantMedia : rawImagesList.length > 0 ? rawImagesList : [img];

	// Extract unique colors from variants (kept for legacy consumers)
	const colors: { name: string; class: string; variantIndex: number }[] = [];
	const seenColors = new Set<string>();
	if (Array.isArray(raw.variants)) {
		raw.variants.forEach((v: RawProductVariant, idx: number) => {
			const colorNameRaw = v.attributes?.color ?? v.color;
			const colorName =
				colorNameRaw === undefined || colorNameRaw === null ? "" : String(colorNameRaw);
			if (colorName && !seenColors.has(colorName)) {
				seenColors.add(colorName);
				colors.push({
					name: colorName,
					class: `bg-gray-400`, // fallback swatch color
					variantIndex: idx,
				});
			}
		});
	}

	// Build generic variation axes. Products with declared axes use those keys;
	// legacy rows fall back to keys shared by every child.
	const axisLabels: Record<string, string> = {
		color: "Color",
		colour: "Color",
		size: "Size",
		storage: "Storage",
		ram: "RAM",
		memory: "Memory",
		capacity: "Capacity",
		material: "Material",
		style: "Style",
		band: "Band",
		finish: "Finish",
		model: "Model",
	};
	const axisLabel = (key: string) =>
		axisLabels[key] || key.replace(/[_-]+/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());

	const rawVariants = Array.isArray(raw.variants) ? raw.variants : [];
	let axisKeys: string[] = [];
	const declaredAxes = raw.variationAxes ?? p.variationAxes;
	if (Array.isArray(declaredAxes) && declaredAxes.length > 0) {
		axisKeys = declaredAxes.filter((k: string) => typeof k === "string" && k.trim() !== "");
	} else if (rawVariants.length > 1) {
		const keySets = rawVariants.map((v: RawProductVariant) =>
			v.attributes && typeof v.attributes === "object" ? Object.keys(v.attributes) : []
		);
		axisKeys = keySets[0].filter((key: string) =>
			keySets.every((keys: string[]) => keys.includes(key))
		);
	}

	const variationAxes = axisKeys
		.map((key) => {
			const values: string[] = [];
			for (const v of rawVariants) {
				const value = v.attributes?.[key] ?? v[key];
				if (value === undefined || value === null || value === "") continue;
				const str = String(value);
				if (!values.includes(str)) values.push(str);
			}
			return { key, label: axisLabel(key), values };
		})
		.filter((axis) => axis.values.length > 0);

	// Extract stock info from first variant
	const firstVariant =
		Array.isArray(raw.variants) && raw.variants.length > 0 ? raw.variants[0] : null;
	const inStock = firstVariant?.stock?.isInStock ?? raw.inStock ?? true;
	const stockCount = firstVariant?.stock?.availableQuantity ?? raw.stockCount ?? 0;

	// Extract category — backend list/detail may provide category object { id, name } or string or categories array
	const firstCategory = Array.isArray(raw.categories) ? raw.categories[0] : undefined;
	const firstCategoryName =
		typeof firstCategory === "string" ? firstCategory : firstCategory?.name || "";
	const firstCategoryId = typeof firstCategory !== "string" ? firstCategory?.id : undefined;

	const categoryName =
		typeof raw.category === "string"
			? raw.category
			: raw.category?.name || firstCategoryName || "";
	const categoryId =
		(typeof raw.category !== "string" ? raw.category?.id : undefined) ||
		firstCategoryId ||
		"";

	const currency = typeof raw.currency === "string" ? raw.currency : "";

	const variants = Array.isArray(raw.variants)
		? raw.variants.map((v: RawProductVariant) => {
				// A variant explicitly priced at 0 must stay 0 (free) — only a
				// missing/non-numeric price falls back to the product price.
				const parsedPrice =
					v.extractedPrice != null
						? Number(v.extractedPrice)
						: typeof v.price === "number"
							? v.price
							: v.price != null && String(v.price).trim() !== ""
								? parseFloat(String(v.price).replace(/[^0-9.-]+/g, ""))
								: NaN;
				const vPrice = Number.isFinite(parsedPrice) ? parsedPrice : priceNum;
				const vOldPrice =
					v.extractedOriginalPrice != null
						? Number(v.extractedOriginalPrice)
						: v.originalPrice || v.oldPrice
							? parseFloat(String(v.originalPrice || v.oldPrice).replace(/[^0-9.-]+/g, ""))
							: undefined;
				const vThumbRaw =
					typeof v.thumbnail === "string"
						? v.thumbnail
						: v.thumbnail?.url || v.thumbnail?.publicUrl || undefined;
				const vThumb = vThumbRaw ? normalizeMediaUrl(vThumbRaw) : undefined;
				const vMedia = Array.isArray(v.media)
					? v.media.map((m: RawProductMedia) => {
							const url = normalizeMediaUrl(
								typeof m === "string" ? m : m?.url || m?.publicUrl || m?.object?.publicUrl
							);
							return typeof m === "object" ? { ...m, url } : { url };
						})
					: [];
				return {
					...v,
					thumbnail: vThumb,
					media: vMedia,
					price: vPrice,
					extractedPrice: vPrice,
					oldPrice: vOldPrice,
					extractedOriginalPrice: vOldPrice,
					originalPrice:
						v.originalPrice || (vOldPrice !== undefined ? formatMoney(vOldPrice, currency) : undefined),
				};
			})
		: [];

	// A bundle's contents and their availability, present only for kits. Each
	// component thumbnail is normalized so the gallery and the contents list can
	// use the same URLs as regular product media.
	const rawBundle = p.bundle ?? raw.bundle;
	const bundleItems: RawBundleItem[] = Array.isArray(rawBundle?.items) ? rawBundle.items : [];
	const bundle =
		rawBundle && bundleItems.length > 0
			? {
					itemCount: Number(rawBundle.itemCount ?? bundleItems.length) || bundleItems.length,
					availableQuantity: Math.max(0, Number(rawBundle.availableQuantity ?? 0) || 0),
					isInStock: Boolean(rawBundle.isInStock),
					items: bundleItems.map((item) => ({
						id: String(item.id ?? ""),
						title: typeof item.title === "string" ? item.title : "",
						slug: typeof item.slug === "string" ? item.slug : "",
						thumbnail: item.thumbnail ? normalizeMediaUrl(item.thumbnail) : undefined,
						quantity: Math.max(1, Number(item.quantity ?? 1) || 1),
						variant: item.variant ? String(item.variant) : undefined,
						availableQuantity: Math.max(0, Number(item.availableQuantity ?? 0) || 0),
						isInStock: Boolean(item.isInStock),
					})),
				}
			: undefined;

	const variantId = raw.variantId || p.variantId || undefined;
	const defaultVariantId = raw.defaultVariantId || raw.variantId || p.variantId || undefined;

	const rawBrand = p.brand ?? raw.brand;
	const brandName =
		typeof rawBrand === "string"
			? rawBrand
			: p.brandName || String(rawBrand?.name ?? "") || "";
	const brandObj =
		typeof rawBrand === "object" && rawBrand !== null
			? {
					...rawBrand,
					logoUrl:
						rawBrand.logoUrl || rawBrand.logo
							? normalizeMediaUrl(rawBrand.logoUrl || rawBrand.logo)
							: undefined,
					logo:
						rawBrand.logoUrl || rawBrand.logo
							? normalizeMediaUrl(rawBrand.logoUrl || rawBrand.logo)
							: undefined,
				}
			: null;

	return {
		...raw,
		id,
		productId: id,
		wishlistItemId,
		variantId,
		defaultVariantId,
		slug,
		product,
		title: titleStr,
		img,
		images,
		colors,
		variationAxes,
		variants,
		price: priceNum,
		extractedPrice: priceNum,
		oldPrice: oldPriceNum,
		extractedOriginalPrice: oldPriceNum,
		originalPrice:
			p.originalPrice ||
			raw.originalPrice ||
			(oldPriceNum !== undefined ? formatMoney(oldPriceNum, currency) : undefined),
		discountPercentage:
			typeof discountPercentage === "number" ? `${discountPercentage}%` : discountPercentage,
		stars,
		reviews,
		inStock,
		stockCount,
		brand: brandName,
		brandObj,
		category: categoryName || p.categoryName || "",
		categoryId,
		bundle,
	};
}
