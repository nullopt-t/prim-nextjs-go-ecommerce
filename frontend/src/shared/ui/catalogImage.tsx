"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import { normalizeMediaUrl, PLACEHOLDER_PRODUCT_IMAGE } from "@/shared/lib/media";

interface CatalogImageProps {
	src?: string;
	alt?: string;
	fill?: boolean;
	width?: number;
	height?: number;
	className?: string;
	priority?: boolean;
	loading?: "eager" | "lazy";
	style?: React.CSSProperties;
	onLoad?: (e: React.SyntheticEvent<HTMLImageElement>) => void;
	onError?: (e: React.SyntheticEvent<HTMLImageElement, Event>) => void;
}

/**
 * Catalog image built on `next/image` so storefront product/review thumbnails
 * get native lazy loading, `srcset`/`sizes` and stable layout-shift handling.
 *
 * A custom loader normalizes media URLs through `normalizeMediaUrl`, so images
 * resolve identically on the server and in the browser regardless of whether
 * media is served from MinIO (`NEXT_PUBLIC_API_URL:9000`), `NEXT_PUBLIC_MEDIA_BASE_URL`, or
 * an absolute CDN origin. This avoids restricting `next.config.mjs`
 * `remotePatterns` to known hosts.
 *
 * When `fill` is set, the parent element must be `position: relative` with
 * defined dimensions (all catalog thumbnail containers already are).
 */
export function CatalogImage({
	src,
	alt = "",
	fill = false,
	width,
	height,
	className = "",
	priority = false,
	loading = "lazy",
	style,
	onLoad,
	onError,
	...rest
}: CatalogImageProps) {
	const [currentSrc, setCurrentSrc] = useState(src ?? PLACEHOLDER_PRODUCT_IMAGE);
	const [failed, setFailed] = useState(false);

	useEffect(() => {
		setCurrentSrc(src ?? PLACEHOLDER_PRODUCT_IMAGE);
		setFailed(false);
	}, [src]);

	// Graceful fallback for unreachable MinIO links: swap to the local
	// placeholder instead of a broken icon / layout-breaking alt text.
	const handleError = (e: React.SyntheticEvent<HTMLImageElement, Event>) => {
		if (!failed && currentSrc !== PLACEHOLDER_PRODUCT_IMAGE) {
			setFailed(true);
			setCurrentSrc(PLACEHOLDER_PRODUCT_IMAGE);
		}
		if (onError) onError(e);
	};

	if (!src) return null;

	return (
		<Image
			loader={({ src: loaderSrc }: { src: string }) =>
				normalizeMediaUrl(loaderSrc)
			}
			src={currentSrc}
			alt={alt}
			fill={fill}
			width={width}
			height={height}
			className={className}
			priority={priority}
			loading={loading}
			style={style}
			onLoad={onLoad}
			onError={handleError}
			{...rest}
		/>
	);
}
