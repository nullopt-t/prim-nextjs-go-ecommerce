"use client";

import { useState, useEffect, useRef } from "react";
import { normalizeMediaUrl } from "@/shared/lib/media";

interface LazyImageProps extends React.ImgHTMLAttributes<HTMLImageElement> {
  src?: string;
  alt?: string;
  containerClassName?: string;
  imageClassName?: string;
}

export function LazyImage({
  src,
  alt = "",
  containerClassName = "",
  imageClassName = "",
  onError,
  ...props
}: LazyImageProps) {
  const [isLoaded, setIsLoaded] = useState(false);
  const normalized = normalizeMediaUrl(src);
  const [imgSrc, setImgSrc] = useState(normalized);
  const imgRef = useRef<HTMLImageElement>(null);

  useEffect(() => {
    setImgSrc(normalizeMediaUrl(src));
    setIsLoaded(false);
  }, [src]);

  // A cached image can complete before React attaches the load handler,
  // which would leave the image stuck invisible. Check completion after
  // every render pass.
  useEffect(() => {
    const img = imgRef.current;
    if (img && img.complete && img.naturalWidth > 0) {
      setIsLoaded(true);
    }
  }, [imgSrc]);

  const handleError = (e: React.SyntheticEvent<HTMLImageElement, Event>) => {
    if (imgSrc !== "/placeholder-product.png") {
      setImgSrc("/placeholder-product.png");
    }
    setIsLoaded(true);
    if (onError) onError(e);
  };

  return (
    <div className={`relative overflow-hidden bg-secondary/50 ${containerClassName}`}>
      {!isLoaded && <div className="absolute inset-0 bg-muted animate-pulse" />}
      {imgSrc && (
        <img
          ref={imgRef}
          src={imgSrc}
          alt={alt}
          className={`transition-opacity duration-500 ${
            isLoaded ? "opacity-100" : "opacity-0"
          } ${imageClassName}`}
          onLoad={() => setIsLoaded(true)}
          onError={handleError}
          {...props}
        />
      )}
    </div>
  );
}
