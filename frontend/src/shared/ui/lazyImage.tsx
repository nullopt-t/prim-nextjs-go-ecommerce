"use client";

import { useState, useEffect } from "react";
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

  useEffect(() => {
    setImgSrc(normalizeMediaUrl(src));
    setIsLoaded(false);
  }, [src]);

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
