import { Link, useRouter } from "@/i18n/navigation";
import { useParams, useSearchParams } from "next/navigation";
import { useTranslations, useLocale } from "next-intl";
import { ChevronRight, Package } from "lucide-react";
import { useState, useMemo, useEffect } from "react";
import { AnimatedSection, CustomButton } from "@/components/ui";
import { ProductsGrid } from "@/components/ui/productsGrid";
import SectionTitle from "@/features/home/components/ui/sectionTitle";
import { useAllProducts, useProductBySlug } from "@/hooks/useCatalog";
import { useCartContext } from "@/context/CartContext";
import { catalogService } from "@/services/catalog";
import { wishlistService } from "@/services/wishlist";
import { toast } from "sonner";

import {
  ProductReview,
  getProductDisplayName,
  resolveProductStock,
  resolveProductImages,
} from "../domain/productDomain";
import { ProductGallery } from "./ui/ProductGallery";
import { ProductActions } from "./ui/ProductActions";
import { ProductSpecs } from "./ui/ProductSpecs";
import { ProductReviews } from "./ui/ProductReviews";

export function ProductDetails() {
  const { id } = useParams();
  const searchParams = useSearchParams();
  const navigate = useRouter();
  const locale = useLocale();
  const t = useTranslations("common");

  const { products: allProducts, loading: allLoading } = useAllProducts();
  const { product, loading: productLoading } = useProductBySlug(id);
  const { addToCart, openCart } = useCartContext();

  const [activeImage, setActiveImage] = useState(0);
  const [selectedColor, setSelectedColor] = useState(0);
  const [quantity, setQuantity] = useState(1);
  const [isAdding, setIsAdding] = useState(false);
  const [isWishlisted, setIsWishlisted] = useState(false);
  const [reviews, setReviews] = useState<ProductReview[]>([]);

  // Synchronize variant/color from URL search params (e.g. ?variant=... or ?color=...)
  useEffect(() => {
    if (!product) return;

    const variantParam = searchParams.get("variant");
    const colorParam = searchParams.get("color");

    if (variantParam && Array.isArray(product.variants)) {
      const vIndex = product.variants.findIndex(
        (v: any) => v.id === variantParam || v.sku === variantParam
      );
      if (vIndex !== -1) {
        // Find if this variant corresponds to a color option
        const matchingColorIdx = (product.colors || []).findIndex(
          (c: any) => c.variantIndex === vIndex
        );
        if (matchingColorIdx !== -1) {
          setSelectedColor(matchingColorIdx);
        }
        return;
      }
    }

    if (colorParam && Array.isArray(product.colors)) {
      const cIndex = product.colors.findIndex(
        (c: any) => c.name.toLowerCase() === colorParam.toLowerCase()
      );
      if (cIndex !== -1) {
        setSelectedColor(cIndex);
      }
    }
  }, [product, searchParams]);

  useEffect(() => {
    window.scrollTo(0, 0);
  }, [id]);

  useEffect(() => {
    let isMounted = true;
    const cleanSlug = Array.isArray(id) ? id[0] : id;
    if (!cleanSlug) return;

    catalogService
      .getProductReviews(cleanSlug)
      .then((res: any) => {
        if (!isMounted) return;
        const list = Array.isArray(res) ? res : Array.isArray(res?.data) ? res.data : [];
        if (list.length > 0) {
          const mapped: ProductReview[] = list.map((r: any) => ({
            id: r.id,
            author: r.user?.name || "Customer",
            avatar: r.user?.avatar || null,
            rating: r.rating || 5,
            date: r.createdAt
              ? new Date(r.createdAt).toLocaleDateString("en-US", {
                  year: "numeric",
                  month: "long",
                  day: "numeric",
                })
              : "Recently",
            title: r.title || "Customer Review",
            content: r.body || r.content || "",
            helpful: r.helpful || 0,
            verified: true,
          }));
          setReviews(mapped);
        }
      })
      .catch((err) => {
        console.error("Failed to load product reviews:", err);
      });

    return () => {
      isMounted = false;
    };
  }, [id]);

  const similarProducts = useMemo(() => {
    if (!product || !allProducts) return [];
    return allProducts
      .filter((p) => p.category === product.category && p.id !== product.id)
      .slice(0, 5);
  }, [product, allProducts]);

  if (allLoading || productLoading) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center text-muted-foreground">
        <div className="size-8 rounded-full border-2 border-accent-brand border-t-transparent animate-spin mb-3" />
        <span className="text-sm">Loading product details...</span>
      </div>
    );
  }

  if (!product) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <Package className="size-12 text-muted-foreground mb-3 opacity-50" />
        <h1 className="text-xl font-bold text-foreground mb-2">Product Not Found</h1>
        <p className="text-muted-foreground text-sm mb-6 max-w-sm">
          The product you are looking for doesn't exist or has been removed.
        </p>
        <Link href="/products">
          <CustomButton text="Browse All Products" className="px-6 py-2.5 text-sm" />
        </Link>
      </div>
    );
  }

  const productName = getProductDisplayName(product, locale);
  const { activeVariant, availableStock, isVariantInStock } = resolveProductStock(product, selectedColor);
  const images = resolveProductImages(product, activeVariant);

  // When selected color changes, reset active image to 0 to show the new variant's main photo
  const handleSelectColor = (idx: number) => {
    setSelectedColor(idx);
    setActiveImage(0);
  };

  const currentPrice = activeVariant?.price ? activeVariant.price / 100 : product.price;
  const currentOldPrice = product.oldPrice;

  const handleAddToCart = async () => {
    if (!isVariantInStock) return;
    setIsAdding(true);
    const variantId = activeVariant?.id || String(product.id);
    const payload = {
      id: String(product.id),
      variantId,
      productName: product.product,
      productPrice: `$${currentPrice}`,
      img: images[0],
      quantity,
      color: (product.colors || [])[selectedColor]?.name || "Default",
    };
    const res = await addToCart(payload);
    setIsAdding(false);
    if (res.success) {
      toast.success(`Added ${quantity}x ${productName} to your cart!`);
      openCart();
    } else {
      toast.error(res.error || "Failed to add to cart");
    }
  };

  const handleBuyNow = async () => {
    if (!isVariantInStock) return;
    setIsAdding(true);
    const payload = {
      id: String(product.id),
      productName: product.product,
      productPrice: `$${product.price}`,
      img: images[0],
      quantity,
      color: (product.colors || [])[selectedColor]?.name || "Default",
    };
    const res = await addToCart(payload);
    setIsAdding(false);
    if (res.success) {
      navigate.push("/checkout");
    } else {
      toast.error(res.error || "Failed to initiate checkout");
    }
  };

  const handleToggleWishlist = async () => {
    try {
      if (!isWishlisted) {
        await wishlistService.addItem({
          id: String(product.id),
          productName: product.product,
          productPrice: `$${product.price}`,
          img: images[0],
        });
        setIsWishlisted(true);
        toast.success(`Saved ${productName} to your wishlist!`);
      } else {
        await wishlistService.removeItem(product.id);
        setIsWishlisted(false);
        toast.info(`Removed ${productName} from wishlist`);
      }
    } catch {
      toast.error("Failed to update wishlist");
    }
  };

  const handleShare = () => {
    if (navigator.clipboard) {
      navigator.clipboard.writeText(window.location.href);
      toast.success("Product link copied to clipboard!");
    } else {
      toast.info("Share URL: " + window.location.href);
    }
  };

  const handleAddReview = (newReviewData: Omit<ProductReview, "id" | "date" | "helpful" | "verified">) => {
    const newRev: ProductReview = {
      ...newReviewData,
      id: Date.now(),
      date: `Reviewed on ${new Date().toLocaleDateString("en-US", {
        year: "numeric",
        month: "long",
        day: "numeric",
      })}`,
      helpful: 0,
      verified: true,
    };
    setReviews((prev) => [newRev, ...prev]);
  };

  const handleVoteHelpful = (reviewId: string | number) => {
    setReviews((prev) =>
      prev.map((r) => {
        if (r.id === reviewId) {
          const isVoted = r.userVoted;
          return {
            ...r,
            helpful: isVoted ? r.helpful - 1 : r.helpful + 1,
            userVoted: !isVoted,
          };
        }
        return r;
      })
    );
  };

  return (
    <AnimatedSection className="w-full">
      {/* Breadcrumb */}
      <nav className="flex items-center gap-1.5 text-xs sm:text-sm text-muted-foreground mb-4 sm:mb-6 overflow-x-auto whitespace-nowrap pb-1 scrollbar-none">
        <Link href="/" className="hover:text-foreground transition-colors">
          Home
        </Link>
        <ChevronRight className="size-3.5 rtl:rotate-180 shrink-0" />
        <Link href="/products" className="hover:text-foreground transition-colors">
          Products
        </Link>
        <ChevronRight className="size-3.5 rtl:rotate-180 shrink-0" />
        <Link
          href={`/products?category=${product.categoryId || product.category}`}
          className="hover:text-foreground transition-colors capitalize"
        >
          {product.category?.replace?.("-", " ") || product.category}
        </Link>
        <ChevronRight className="size-3.5 rtl:rotate-180 shrink-0" />
        <span className="text-foreground font-medium truncate max-w-[200px]">{productName}</span>
      </nav>

      {/* Main Product Layout */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-10 lg:gap-16">
        <ProductGallery
          images={images}
          productName={productName}
          activeImage={activeImage}
          onSelectImage={setActiveImage}
          discountPercentage={product.discountPercentage}
        />

        <ProductActions
          productCategory={product.category}
          productName={productName}
          stars={product.stars}
          reviewsCount={product.reviews}
          price={currentPrice}
          oldPrice={currentOldPrice}
          currency={t("currency")}
          colors={product.colors}
          selectedColor={selectedColor}
          onSelectColor={handleSelectColor}
          availableStock={availableStock}
          isVariantInStock={isVariantInStock}
          quantity={quantity}
          onQuantityChange={setQuantity}
          isAdding={isAdding}
          isWishlisted={isWishlisted}
          onAddToCart={handleAddToCart}
          onBuyNow={handleBuyNow}
          onToggleWishlist={handleToggleWishlist}
          onShare={handleShare}
        />
      </div>

      {/* Specifications & Shipping Sections */}
      <ProductSpecs productName={productName} category={product.category} />

      {/* Reviews Section */}
      <div className="mt-16">
        <ProductReviews
          stars={product.stars}
          reviewsCount={product.reviews}
          reviews={reviews}
          onAddReview={handleAddReview}
          onVoteHelpful={handleVoteHelpful}
        />
      </div>

      {/* Similar Products */}
      {similarProducts.length > 0 && (
        <div className="mt-8 sm:mt-12 mb-6 sm:mb-8">
          <SectionTitle
            title="Similar Products"
            link={`/products?category=${product.categoryId || product.category}`}
          />
          <ProductsGrid
            products={similarProducts}
            className="grid-cols-2 sm:grid-cols-3 lg:grid-cols-5"
          />
        </div>
      )}
    </AnimatedSection>
  );
}
