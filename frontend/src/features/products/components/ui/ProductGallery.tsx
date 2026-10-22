import { useTranslations } from "next-intl";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { LazyImage } from "@/components/ui/lazyImage";

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

  const goToPrevious = () => {
    onSelectImage((currentIndex - 1 + safeImages.length) % safeImages.length);
  };

  const goToNext = () => {
    onSelectImage((currentIndex + 1) % safeImages.length);
  };

  return (
    <div className="relative">
      {/* Vertical thumbnail rail */}
      {hasThumbnails && (
        <div className="absolute inset-y-0 start-0 flex w-16 flex-col gap-2.5 overflow-y-auto pe-1 sm:w-20 lg:w-24 scrollbar-thin">
          {safeImages.map((img, i) => {
            const isActive = i === currentIndex;
            return (
              <button
                key={`${img}-${i}`}
                type="button"
                onClick={() => onSelectImage(i)}
                className={`relative aspect-square w-full shrink-0 overflow-hidden rounded-xl border bg-card transition-all cursor-pointer ${
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

      {/* Main image */}
      <div className={hasThumbnails ? "ms-20 sm:ms-24 lg:ms-28" : ""}>
        <div className="group relative aspect-square w-full overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
          <LazyImage
            key={currentImage}
            src={currentImage}
            alt={productName}
            containerClassName="w-full h-full bg-card"
            imageClassName="object-contain object-center w-full h-full p-4 sm:p-7"
          />

          {discountPercentage && (
            <span className="absolute top-3 start-3 rounded-full bg-destructive px-3 py-1 text-xs font-bold text-white shadow-sm sm:top-4 sm:start-4 sm:text-sm">
              {t("gallery.off", { discount: discountPercentage })}
            </span>
          )}

          {hasThumbnails && (
            <>
              <button
                type="button"
                onClick={goToPrevious}
                aria-label={t("gallery.previousImage")}
                className="absolute start-3 top-1/2 flex size-9 -translate-y-1/2 cursor-pointer items-center justify-center rounded-full border border-border bg-background/90 text-foreground opacity-100 shadow-sm backdrop-blur-sm transition-all hover:bg-background focus:opacity-100 sm:size-10 sm:opacity-0 sm:group-hover:opacity-100"
              >
                <ChevronLeft className="size-4 rtl:rotate-180" />
              </button>
              <button
                type="button"
                onClick={goToNext}
                aria-label={t("gallery.nextImage")}
                className="absolute end-3 top-1/2 flex size-9 -translate-y-1/2 cursor-pointer items-center justify-center rounded-full border border-border bg-background/90 text-foreground opacity-100 shadow-sm backdrop-blur-sm transition-all hover:bg-background focus:opacity-100 sm:size-10 sm:opacity-0 sm:group-hover:opacity-100"
              >
                <ChevronRight className="size-4 rtl:rotate-180" />
              </button>
              <span
                dir="ltr"
                className="absolute bottom-3 end-3 rounded-full bg-foreground/75 px-2.5 py-1 text-[11px] font-semibold text-background backdrop-blur-sm"
              >
                {currentIndex + 1} / {safeImages.length}
              </span>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
