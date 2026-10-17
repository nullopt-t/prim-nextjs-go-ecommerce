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
import { useWishlist } from "@/hooks/useWishlist";
import { notify, apiErrorMessage } from "@/lib/notify";

import { ProductReview } from "../types";
import {
  getProductDisplayName,
  resolveProductStock,
  resolveProductImages,
} from "../utils/productHelpers";
import { ProductGallery } from "./ui/ProductGallery";
import { ProductActions } from "./ui/ProductActions";
import { ProductSpecs } from "./ui/ProductSpecs";
import { ProductReviews } from "./ui/ProductReviews";

import { useAuthContext } from "@/context/AuthContext";
import { userService } from "@/services/user";

export function ProductDetails() {
  const { id } = useParams();
  const searchParams = useSearchParams();
  const navigate = useRouter();
  const locale = useLocale();
  const t = useTranslations("common");
  const tReviews = useTranslations("reviews");
  const tProducts = useTranslations("products");
  const { user } = useAuthContext();

  const { products: allProducts, loading: allLoading } = useAllProducts();
  const { product, loading: productLoading } = useProductBySlug(id);
  const { addToCart } = useCartContext();
  const { isWishlisted: checkIsWishlisted, toggleWishlist } = useWishlist();

  const [activeImage, setActiveImage] = useState(0);
  const [selectedColor, setSelectedColor] = useState(0);
  const [quantity, setQuantity] = useState(1);
  const [isAdding, setIsAdding] = useState(false);
  const [reviews, setReviews] = useState<ProductReview[]>([]);
  const [canReview, setCanReview] = useState(false);

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
      .then((publicRes: any) => {
        if (!isMounted) return;
        const publicList = Array.isArray(publicRes) ? publicRes : Array.isArray(publicRes?.data) ? publicRes.data : [];

        const mapped: ProductReview[] = publicList.map((r: any) => {
          const isOwn = user?.id && String(r.userId) === String(user.id);
          return {
            id: r.id,
            author: isOwn ? (user.name || tProducts("details.you")) : r.user?.name || tProducts("details.customer"),
            avatar: isOwn ? null : r.user?.avatar || null,
            rating: r.rating || 5,
            date: r.createdAt
              ? new Date(r.createdAt).toLocaleDateString("en-US", {
                  year: "numeric",
                  month: "long",
                  day: "numeric",
                })
              : tProducts("details.recently"),
            title: r.title || tProducts("details.customerReview"),
            content: r.body || r.content || "",
            helpful: r.helpful || 0,
            verified: true,
          };
        });

        setReviews(mapped);
      })
      .catch((err) => {
        console.error("Failed to load product reviews:", err);
      });

    return () => {
      isMounted = false;
    };
  }, [id, user, product?.id]);

  // Check if current logged-in user is eligible to review this product via dedicated endpoint
  useEffect(() => {
    let isMounted = true;
    if (!user || !product?.id) {
      setCanReview(false);
      return;
    }

    catalogService
      .checkReviewEligibility(String(product.id))
      .then((res: any) => {
        if (!isMounted) return;
        const eligible = Boolean(res?.data?.canReview ?? res?.canReview);
        setCanReview(eligible);
      })
      .catch(() => {
        if (isMounted) setCanReview(false);
      });

    return () => {
      isMounted = false;
    };
  }, [user, product?.id]);

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
        <span className="text-sm">{tProducts("details.loading")}</span>
      </div>
    );
  }

  if (!product) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <Package className="size-12 text-muted-foreground mb-3 opacity-50" />
        <h1 className="text-xl font-bold text-foreground mb-2">{tProducts("details.notFound")}</h1>
        <p className="text-muted-foreground text-sm mb-6 max-w-sm">
          {tProducts("details.notFoundDescription")}
        </p>
        <Link href="/products">
          <CustomButton text={tProducts("details.browseAll")} className="px-6 py-2.5 text-sm" />
        </Link>
      </div>
    );
  }

  const productName = getProductDisplayName(product, locale);
  const { activeVariant, availableStock, isVariantInStock } = resolveProductStock(product, selectedColor);
  const images = resolveProductImages(product, activeVariant);

  const isWishlisted = Boolean(
    activeVariant?.id
      ? checkIsWishlisted(String(product?.id), activeVariant.id) || checkIsWishlisted(String(product?.slug), activeVariant.id)
      : (product?.id && checkIsWishlisted(String(product.id))) || (product?.slug && checkIsWishlisted(String(product.slug)))
  );

  // When selected color changes, reset active image to 0 to show the new variant's main photo
  const handleSelectColor = (idx: number) => {
    setSelectedColor(idx);
    setActiveImage(0);
  };

  const resolvePrice = (val: any): number => {
    if (typeof val === "number" && !isNaN(val)) return val;
    if (!val) return 0;
    const parsed = parseFloat(String(val).replace(/[^0-9.-]+/g, ""));
    return isNaN(parsed) ? 0 : parsed;
  };

  const activeVariantPrice = (activeVariant as any)?.extractedPrice ?? activeVariant?.price;
  const numVariantPrice = resolvePrice(activeVariantPrice);
  const numProductPrice = resolvePrice((product as any)?.extractedPrice ?? product.price);
  const currentPrice = numVariantPrice > 0 ? numVariantPrice : numProductPrice;

  const activeVariantOldPrice = (activeVariant as any)?.extractedOriginalPrice ?? (activeVariant as any)?.oldPrice ?? (activeVariant as any)?.originalPrice;
  const numVariantOldPrice = resolvePrice(activeVariantOldPrice);
  const numProductOldPrice = resolvePrice((product as any)?.extractedOriginalPrice ?? product.oldPrice ?? (product as any)?.originalPrice);
  const rawOldPrice = numVariantOldPrice > 0 ? numVariantOldPrice : numProductOldPrice;
  const currentOldPrice = rawOldPrice > currentPrice ? rawOldPrice : undefined;

  const calculatedDiscount = currentOldPrice && currentOldPrice > currentPrice
    ? `${Math.round(((currentOldPrice - currentPrice) / currentOldPrice) * 100)}%`
    : undefined;
  const discountPercentage = product.discountPercentage || calculatedDiscount;

  const handleAddToCart = async () => {
    if (!isVariantInStock) return;
    setIsAdding(true);
    const resolvedVariantId =
      activeVariant?.id ||
      (product.variants && product.variants.find((v: any) => v.isDefault)?.id) ||
      (product.variants && product.variants[0]?.id) ||
      product.variantId ||
      product.defaultVariantId;
    const payload = {
      id: String(product.id),
      slug: product.slug,
      variantId: resolvedVariantId,
      productName: product.product,
      productPrice: `EGP ${currentPrice}`,
      img: images[0],
      quantity,
      color: (product.colors || [])[selectedColor]?.name || "Default",
    };
    const res = await addToCart(payload);
    setIsAdding(false);
    if (res.success) {
      notify.success(tProducts("details.addedToCart", { quantity, productName }));
    } else {
      notify.error(apiErrorMessage(res, tProducts("details.addFailed")));
    }
  };

  const handleBuyNow = async () => {
    if (!isVariantInStock) return;
    setIsAdding(true);
    const resolvedVariantId =
      activeVariant?.id ||
      (product.variants && product.variants.find((v: any) => v.isDefault)?.id) ||
      (product.variants && product.variants[0]?.id) ||
      product.variantId ||
      product.defaultVariantId;
    const payload = {
      id: String(product.id),
      slug: product.slug,
      variantId: resolvedVariantId,
      productName: product.product,
      productPrice: `EGP ${currentPrice}`,
      img: images[0],
      quantity,
      color: (product.colors || [])[selectedColor]?.name || "Default",
    };
    const res = await addToCart(payload);
    setIsAdding(false);
    if (res.success) {
      navigate.push("/checkout");
    } else {
      notify.error(apiErrorMessage(res, tProducts("details.checkoutFailed")));
    }
  };

  const handleToggleWishlist = async () => {
    if (!product?.id) return;
    await toggleWishlist(String(product.id), activeVariant?.id, productName);
  };

  const handleShare = () => {
    if (navigator.clipboard) {
      navigator.clipboard.writeText(window.location.href);
      notify.success(tProducts("details.linkCopied"));
    } else {
      notify.info(tProducts("details.shareUrl", { url: window.location.href }));
    }
  };

  const handleAddReview = async (newReviewData: { rating: number; title: string; content: string }) => {
    if (!product?.id) {
      notify.error(tReviews("missingProduct"));
      return false;
    }

    try {
      const res: any = await catalogService.submitReview({
        productId: String(product.id),
        rating: newReviewData.rating,
        title: newReviewData.title,
        body: newReviewData.content,
      });

      const serverReview = res?.data || res;
      notify.success(tReviews("submitSuccess"));

      const newRev: ProductReview = {
        id: serverReview?.id || Date.now(),
        author: tProducts("details.you"),
        avatar: null,
        rating: newReviewData.rating,
        title: newReviewData.title,
        content: newReviewData.content,
        date: tProducts("details.reviewedOn", {
          date: new Date().toLocaleDateString("en-US", {
            year: "numeric",
            month: "long",
            day: "numeric",
          }),
        }),
        helpful: 0,
        verified: true,
      };

      setReviews((prev) => [newRev, ...prev]);
      setCanReview(false);
      return true;
    } catch (err: any) {
      const errCode = err?.response?.data?.code;
      const statusCode = err?.response?.status;

      if (errCode === "REVIEW_NOT_ELIGIBLE") {
        notify.error(tReviews("errors.notEligible"));
      } else if (errCode === "REVIEW_ALREADY_EXISTS" || statusCode === 409) {
        notify.error(tReviews("errors.alreadyExists"));
      } else if (statusCode === 401) {
        notify.error(tReviews("errors.unauthorized"));
      } else {
        notify.error(tReviews("submitFailed"));
      }
      return false;
    }
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
          {t("header.home")}
        </Link>
        <ChevronRight className="size-3.5 rtl:rotate-180 shrink-0" />
        <Link href="/products" className="hover:text-foreground transition-colors">
          {tProducts("details.products")}
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
      <div className="max-w-7xl mx-auto grid grid-cols-1 md:grid-cols-2 gap-10 lg:gap-16">
        <ProductGallery
          images={images}
          productName={productName}
          activeImage={activeImage}
          onSelectImage={setActiveImage}
          discountPercentage={discountPercentage}
        />

        <ProductActions
          productCategory={product.category}
          productName={productName}
          stars={product.stars}
          reviewsCount={reviews.length || product.reviews || 0}
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
      <ProductSpecs
        productName={productName}
        category={product.category}
        brand={typeof product.brand === "string" ? product.brand : product.brand?.name}
        description={product.description}
        attributes={activeVariant?.attributes}
      />

      {/* Reviews Section */}
      <div className="mt-16">
        <ProductReviews
          stars={product.stars}
          reviewsCount={reviews.length || product.reviews || 0}
          reviews={reviews}
          canReview={canReview}
          onAddReview={handleAddReview}
          onVoteHelpful={handleVoteHelpful}
        />
      </div>

      {/* Similar Products */}
      {similarProducts.length > 0 && (
        <div className="mt-8 sm:mt-12 mb-6 sm:mb-8">
          <SectionTitle
            title={tProducts("details.similarProducts")}
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
