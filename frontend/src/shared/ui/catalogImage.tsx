"use client";

import Image from "next/image";
import { normalizeMediaUrl } from "@/shared/lib/media";

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
	if (!src) return null;

	return (
		<Image
			loader={({ src: loaderSrc }: { src: string }) =>
				normalizeMediaUrl(loaderSrc)
			}
			src={src}
			alt={alt}
			fill={fill}
			width={width}
			height={height}
			className={className}
			priority={priority}
			loading={loading}
			style={style}
			onLoad={onLoad}
			onError={onError}
			{...rest}
		/>
	);
}
