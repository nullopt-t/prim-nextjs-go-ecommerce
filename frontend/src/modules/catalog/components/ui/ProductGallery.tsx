"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useTranslations } from "next-intl";
import { ChevronLeft, ChevronRight, X, ZoomIn, ZoomOut } from "lucide-react";
import { LazyImage } from "@/shared/ui/lazyImage";

interface ProductGalleryProps {
  images: string[];
  productName: string;
  activeImage: number;
  onSelectImage: (index: number) => void;
  discountPercentage?: string;
}

export function ProductGallery({
  images,
  productName,
  activeImage,
  onSelectImage,
  discountPercentage,
}: ProductGalleryProps) {
  const t = useTranslations("products");
  const safeImages = images.length > 0 ? images : ["/placeholder-product.png"];
  const currentIndex = Math.min(Math.max(activeImage, 0), safeImages.length - 1);
  const currentImage = safeImages[currentIndex];
  const hasThumbnails = safeImages.length > 1;

  const trackRef = useRef<HTMLDivElement>(null);
  const dialogRef = useRef<HTMLDialogElement>(null);
  const closeButtonRef = useRef<HTMLButtonElement>(null);
  // Set when the swipe track itself changed the selection, so the sync effect
  // does not scroll the track back to where it already is.
  const scrollDrivenRef = useRef(false);

  const [lightboxOpen, setLightboxOpen] = useState(false);
  const [zoomed, setZoomed] = useState(false);
  const [zoomOrigin, setZoomOrigin] = useState("50% 50%");

  const goToIndex = useCallback(
    (index: number) => {
      const total = safeImages.length;
      scrollDrivenRef.current = false;
      onSelectImage(((index % total) + total) % total);
    },
    [onSelectImage, safeImages.length],
  );

  const goToPrevious = () => {
    setZoomed(false);
    goToIndex(currentIndex - 1);
  };

  const goToNext = () => {
    setZoomed(false);
    goToIndex(currentIndex + 1);
  };

  // Keep the swipe track aligned when the selection changes from a thumbnail,
  // the lightbox, an arrow, or a variant change.
  useEffect(() => {
    const track = trackRef.current;
    if (!track || !track.clientWidth) return;
    if (scrollDrivenRef.current) {
      scrollDrivenRef.current = false;
      return;
    }
    // RTL scroll positions are negative in every engine we target.
    const rtl = getComputedStyle(track).direction === "rtl";
    const left = (rtl ? -1 : 1) * currentIndex * track.clientWidth;
    if (Math.abs(track.scrollLeft - left) > 2) track.scrollTo({ left });
  }, [currentIndex, safeImages.length]);

  // A native modal <dialog> traps focus and closes on Esc for free.
  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;
    if (lightboxOpen) {
      if (!dialog.open) dialog.showModal();
      closeButtonRef.current?.focus();
      const previousOverflow = document.body.style.overflow;
      document.body.style.overflow = "hidden";
      return () => {
        document.body.style.overflow = previousOverflow;
      };
    }
    if (dialog.open) dialog.close();
  }, [lightboxOpen]);

  const handleTrackScroll = () => {
    const track = trackRef.current;
    if (!track || !track.clientWidth || !hasThumbnails) return;
    const index = Math.min(
      Math.max(Math.round(Math.abs(track.scrollLeft) / track.clientWidth), 0),
      safeImages.length - 1,
    );
    if (index === currentIndex) return;
    scrollDrivenRef.current = true;
    onSelectImage(index);
  };

  const handleDialogKeyDown = (event: React.KeyboardEvent<HTMLDialogElement>) => {
    if (!hasThumbnails) return;
    const rtl = getComputedStyle(event.currentTarget).direction === "rtl";
    if (event.key === (rtl ? "ArrowRight" : "ArrowLeft")) {
      event.preventDefault();
      goToPrevious();
    } else if (event.key === (rtl ? "ArrowLeft" : "ArrowRight")) {
      event.preventDefault();
      goToNext();
    }
  };

  const toggleZoom = (event: React.MouseEvent<HTMLElement>) => {
    const rect = event.currentTarget.getBoundingClientRect();
    const x = ((event.clientX - rect.left) / rect.width) * 100;
    const y = ((event.clientY - rect.top) / rect.height) * 100;
    setZoomOrigin(`${x}% ${y}%`);
    setZoomed((previous) => !previous);
  };

  return (
    <div className="relative">
      {/* Vertical thumbnail rail */}
      {hasThumbnails && (
        <div className="absolute inset-y-0 start-0 hidden w-20 flex-col gap-2.5 overflow-y-auto pe-1 md:flex lg:w-24 scrollbar-thin">
          {safeImages.map((img, i) => {
            const isActive = i === currentIndex;
            return (
              <button
                key={`${img}-${i}`}
                type="button"
                onClick={() => onSelectImage(i)}
                className={`relative aspect-square w-full shrink-0 cursor-pointer overflow-hidden rounded-xl border bg-card transition-all ${
                  isActive
                    ? "border-accent-brand ring-2 ring-accent-brand/25 shadow-sm"
                    : "border-border/70 hover:border-accent-brand/60 hover:shadow-sm"
                }`}
                aria-label={t("gallery.selectImage", { index: i + 1 })}
                aria-current={isActive ? "true" : undefined}
              >
                <LazyImage
                  src={img}
                  alt={t("gallery.thumbnailAlt", { productName, index: i + 1 })}
                  containerClassName="w-full h-full bg-card"
                  imageClassName={`object-contain object-center w-full h-full p-1.5 transition-opacity ${
                    isActive ? "opacity-100" : "opacity-70 hover:opacity-100"
                  }`}
                />
              </button>
            );
          })}
        </div>
      )}

      {/* Swipeable main image track (mobile bullets live below it) */}
      <div className={hasThumbnails ? "md:ms-20 lg:ms-24" : ""}>
        <div className="group relative overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
          <div
            ref={trackRef}
            onScroll={handleTrackScroll}
            className="scrollbar-none flex snap-x snap-mandatory overflow-x-auto overscroll-x-contain"
          >
            {safeImages.map((img, i) => (
              <div key={`${img}-${i}`} className="w-full shrink-0 snap-center">
                <button
                  type="button"
                  onClick={() => setLightboxOpen(true)}
                  className="relative block aspect-square w-full cursor-zoom-in overflow-hidden"
                  aria-label={t("gallery.viewFullscreen", { productName })}
                >
                  <LazyImage
                    src={img}
                    alt={i === currentIndex ? productName : t("gallery.thumbnailAlt", { productName, index: i + 1 })}
                    containerClassName="w-full h-full bg-card"
                    imageClassName="object-contain object-center w-full h-full p-4 sm:p-7"
                    loading={i === currentIndex ? "eager" : "lazy"}
                    fetchPriority={i === currentIndex ? "high" : "auto"}
                    draggable={false}
                  />
                </button>
              </div>
            ))}
          </div>

          {discountPercentage && (
            <span className="pointer-events-none absolute top-3 start-3 rounded-full bg-destructive px-3 py-1 text-xs font-bold text-white shadow-sm sm:top-4 sm:start-4 sm:text-sm">
              {t("gallery.off", { discount: discountPercentage })}
            </span>
          )}

          {hasThumbnails && (
            <>
              <button
                type="button"
                onClick={goToPrevious}
                aria-label={t("gallery.previousImage")}
                className="absolute start-3 top-1/2 flex size-11 -translate-y-1/2 cursor-pointer items-center justify-center rounded-full border border-border bg-background/90 text-foreground opacity-100 shadow-sm backdrop-blur-sm transition-all hover:bg-background focus:opacity-100 sm:opacity-0 sm:group-hover:opacity-100"
              >
                <ChevronLeft className="size-4 rtl:rotate-180" />
              </button>
              <button
                type="button"
                onClick={goToNext}
                aria-label={t("gallery.nextImage")}
                className="absolute end-3 top-1/2 flex size-11 -translate-y-1/2 cursor-pointer items-center justify-center rounded-full border border-border bg-background/90 text-foreground opacity-100 shadow-sm backdrop-blur-sm transition-all hover:bg-background focus:opacity-100 sm:opacity-0 sm:group-hover:opacity-100"
              >
                <ChevronRight className="size-4 rtl:rotate-180" />
              </button>
              <span
                dir="ltr"
                className="pointer-events-none absolute bottom-3 start-3 hidden rounded-full bg-foreground/75 px-2.5 py-1 text-[11px] font-semibold text-background backdrop-blur-sm md:inline-flex"
              >
                {currentIndex + 1} / {safeImages.length}
              </span>
            </>
          )}

          {hasThumbnails && (
            <span className="pointer-events-none absolute bottom-3 end-3 flex size-8 items-center justify-center rounded-full bg-foreground/70 text-background backdrop-blur-sm">
              <ZoomIn className="size-4" aria-hidden="true" />
            </span>
          )}
        </div>

        {/* Mobile bullet indicators */}
        {hasThumbnails && (
          <div className="mt-1 flex flex-wrap items-center justify-center gap-1 md:hidden">
            {safeImages.map((img, i) => (
              <button
                key={`bullet-${img}-${i}`}
                type="button"
                onClick={() => goToIndex(i)}
                aria-label={t("gallery.goToSlide", { index: i + 1 })}
                aria-current={i === currentIndex ? "true" : undefined}
                className="flex min-h-11 min-w-11 cursor-pointer items-center justify-center"
              >
                <span
                  aria-hidden="true"
                  className={`block size-2.5 rounded-full transition-colors ${
                    i === currentIndex ? "bg-accent-brand" : "bg-border"
                  }`}
                />
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Fullscreen viewer */}
      <dialog
        ref={dialogRef}
        aria-label={t("gallery.lightboxLabel", { productName })}
        aria-modal="true"
        onCancel={(event) => {
          event.preventDefault();
          setLightboxOpen(false);
        }}
        onClose={() => setLightboxOpen(false)}
        onKeyDown={handleDialogKeyDown}
        onClick={(event) => {
          if (event.target === event.currentTarget) setLightboxOpen(false);
        }}
        className="m-auto w-[calc(100%-1rem)] max-w-5xl overflow-hidden rounded-2xl border border-border bg-card p-0 text-foreground shadow-2xl backdrop:bg-foreground/80 backdrop:backdrop-blur-sm"
      >
        <div className="flex items-center justify-between gap-2 border-b border-border px-2 py-2">
          <button
            ref={closeButtonRef}
            type="button"
            onClick={() => setLightboxOpen(false)}
            aria-label={t("gallery.closeLightbox")}
            className="flex size-11 cursor-pointer items-center justify-center rounded-full text-foreground transition-colors hover:bg-secondary"
          >
            <X className="size-5" aria-hidden="true" />
          </button>

          <button
            type="button"
            onClick={() => {
              setZoomOrigin("50% 50%");
              setZoomed((previous) => !previous);
            }}
            aria-pressed={zoomed}
            aria-label={zoomed ? t("gallery.zoomOut") : t("gallery.zoomIn")}
            className="flex size-11 cursor-pointer items-center justify-center rounded-full text-foreground transition-colors hover:bg-secondary"
          >
            {zoomed ? <ZoomOut className="size-5" aria-hidden="true" /> : <ZoomIn className="size-5" aria-hidden="true" />}
          </button>
        </div>

        <div
          className="relative flex h-[70vh] w-full items-center justify-center overflow-hidden bg-background sm:h-[80vh]"
          onClick={toggleZoom}
        >
          <LazyImage
            key={currentImage}
            src={currentImage}
            alt={productName}
            containerClassName="flex h-full w-full items-center justify-center bg-background"
            imageClassName={`h-full w-full object-contain p-4 ${zoomed ? "cursor-zoom-out" : "cursor-zoom-in"}`}
            style={{
              transform: zoomed ? "scale(2.5)" : "scale(1)",
              transformOrigin: zoomOrigin,
              transition: "transform 300ms ease-out",
            }}
            draggable={false}
          />

          {hasThumbnails && (
            <>
              <button
                type="button"
                onClick={(event) => {
                  event.stopPropagation();
                  goToPrevious();
                }}
                aria-label={t("gallery.previousImage")}
                className="absolute start-3 top-1/2 flex size-11 -translate-y-1/2 cursor-pointer items-center justify-center rounded-full border border-border bg-background/90 text-foreground shadow-sm backdrop-blur-sm transition-colors hover:bg-background"
              >
                <ChevronLeft className="size-5 rtl:rotate-180" />
              </button>
              <button
                type="button"
                onClick={(event) => {
                  event.stopPropagation();
                  goToNext();
                }}
                aria-label={t("gallery.nextImage")}
                className="absolute end-3 top-1/2 flex size-11 -translate-y-1/2 cursor-pointer items-center justify-center rounded-full border border-border bg-background/90 text-foreground shadow-sm backdrop-blur-sm transition-colors hover:bg-background"
              >
                <ChevronRight className="size-5 rtl:rotate-180" />
              </button>
              <span
                dir="ltr"
                aria-live="polite"
                className="pointer-events-none absolute bottom-3 rounded-full bg-foreground/75 px-3 py-1 text-xs font-semibold text-background backdrop-blur-sm"
              >
                {currentIndex + 1} / {safeImages.length}
              </span>
            </>
          )}
        </div>
      </dialog>
    </div>
  );
}
