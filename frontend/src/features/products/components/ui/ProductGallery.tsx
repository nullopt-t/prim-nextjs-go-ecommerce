import { useTranslations } from "next-intl";
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
  const currentImage = images[activeImage] || images[0] || "/placeholder-product.png";

  return (
    <div className="flex flex-col gap-4">
      {/* Main image */}
      <div className="relative aspect-square w-full overflow-hidden rounded-2xl border border-border/50 bg-secondary/30">
        <LazyImage
          src={currentImage}
          alt={productName}
          containerClassName="w-full h-full"
          imageClassName="object-cover object-center w-full h-full"
        />
        {discountPercentage && (
          <span className="absolute top-4 start-4 bg-destructive text-white px-3 py-1 rounded-full text-sm font-semibold shadow-sm">
            {t("gallery.off", { discount: discountPercentage })}
          </span>
        )}
      </div>

      {/* Thumbnails */}
      {images.length > 1 && (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {images.map((img, i) => (
            <button
              key={i}
              type="button"
              onClick={() => onSelectImage(i)}
              className={`aspect-square rounded-xl overflow-hidden border ${
                i === activeImage
                  ? "border-accent-brand ring-2 ring-accent-brand/20"
                  : "border-border/50 hover:border-border"
              } transition-all cursor-pointer`}
              aria-label={t("gallery.selectImage", { index: i + 1 })}
            >
              <LazyImage
                src={img}
                alt={t("gallery.thumbnailAlt", { productName, index: i + 1 })}
                containerClassName={`w-full h-full transition-opacity ${
                  i === activeImage ? "opacity-100" : "opacity-70 hover:opacity-100"
                }`}
                imageClassName="object-cover object-center w-full h-full"
              />
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
