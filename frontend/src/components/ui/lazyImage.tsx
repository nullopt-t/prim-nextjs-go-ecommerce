"use client";

import { useState } from "react";

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
  ...props
}: LazyImageProps) {
  const [isLoaded, setIsLoaded] = useState(false);

  return (
    <div className={`relative overflow-hidden bg-secondary/50 ${containerClassName}`}>
      {!isLoaded && <div className="absolute inset-0 bg-muted animate-pulse" />}
      {src && (
        <img
          src={src}
          alt={alt}
          className={`transition-opacity duration-500 ${
            isLoaded ? "opacity-100" : "opacity-0"
          } ${imageClassName}`}
          onLoad={() => setIsLoaded(true)}
          {...props}
        />
      )}
    </div>
  );
}
