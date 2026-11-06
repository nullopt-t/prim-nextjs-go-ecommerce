"use client";

import { Link, useRouter } from "@/shared/i18n/navigation";import { useParams, useSearchParams } from "next/navigation";
import { useTranslations, useLocale } from "next-intl";
import { ChevronRight, Package } from "lucide-react";
import { useState, useMemo, useEffect } from "react";
import { AnimatedSection, CustomButton } from "@/shared/ui";
import { ProductsGrid } from "@/shared/ui/productsGrid";
import { CardDetails } from "@/shared/ui/productsCard";
import SectionTitle from "@/modules/home/components/ui/sectionTitle";
import { useAllProducts, useProductBySlug } from "@/modules/catalog/model/useCatalog";
import { useCartContext } from "@/modules/cart/model/cartContext";
import { catalogService } from "@/modules/catalog/api/catalogService";
import { useWishlist } from "@/modules/wishlist/model/useWishlist";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { copyTextToClipboard } from "@/shared/lib/clipboard";
import { recordRecentlyViewed } from "@/shared/lib/recentlyViewed";
import { recordProductView } from "@/modules/history/lib/recordProductView";

import type { ProductDetailsEntity, ProductReview, ProductVariant, Review } from "../types";
import {
  getProductDisplayName,
  resolveVariantByAttributes,
  resolveVariantStock,
  resolveProductImageEntries,
} from "../utils/productHelpers";
import { ProductGallery } from "./ui/ProductGallery";
import { ProductActions } from "./ui/ProductActions";
import { ProductSpecs } from "./ui/ProductSpecs";
import { ProductHighlights } from "./ui/ProductHighlights";
import { ProductReviews } from "./ui/ProductReviews";
import { CatalogFeedLoader } from "./CatalogFeedLoader";

import { useAuthContext } from "@/modules/auth/model/authContext";
import { useCurrency } from "@/modules/settings/model/publicSettingsContext";
import { formatMoney } from "@/shared/lib/money";

export function ProductDetails({ initialProduct = null }: { initialProduct?: ProductDetailsEntity | null }) {
  const { id } = useParams();
  const searchParams = useSearchParams();
  const navigate = useRouter();
  const locale = useLocale();
  const t = useTranslations("common");
  const currency = useCurrency();
  const tReviews = useTranslations("reviews");
  const tProducts = useTranslations("products");
  const { user } = useAuthContext();

  const { products: allProducts, loading: allLoading } = useAllProducts();
  const { product, loading: productLoading } = useProductBySlug(id, initialProduct);
  const { addToCart } = useCartContext();
  const { isWishlisted: checkIsWishlisted, toggleWishlist } = useWishlist();

  const [activeImage, setActiveImage] = useState(0);
  const [selectedValues, setSelectedValues] = useState<Record<string, string>>({});
  const [quantity, setQuantity] = useState(1);
  const [isAdding, setIsAdding] = useState(false);
  const [reviews, setReviews] = useState<ProductReview[]>([]);
  const [canReview, setCanReview] = useState(false);

  // Listen for real-time stock updates to immediately update availableStock & isVariantInStock
  useEffect(() => {
    const handleStockUpdate = (e: Event) => {
      const customEvent = e as CustomEvent;
      const detail = customEvent.detail;
      if (!detail?.variantId || !product?.variants) return;

      const variantIndex = product.variants.findIndex((v: ProductVariant) => v.id === detail.variantId);
      if (variantIndex !== -1) {
        const updatedVariants = [...product.variants];
        const v = updatedVariants[variantIndex];
        updatedVariants[variantIndex] = {
          ...v,
          stock: {
            ...v.stock,
            availableQuantity: detail.currentStock,
            isInStock: detail.currentStock > 0,
          },
        };
        product.variants = updatedVariants;
        // Trigger re-render by syncing state
        setSelectedValues((prev) => ({ ...prev }));
      }
    };

    window.addEventListener("admin-stock-updated", handleStockUpdate);
    return () => {
      window.removeEventListener("admin-stock-updated", handleStockUpdate);
    };
  }, [product]);

  // Initialize the matrix selection from a URL-pinned variant (or the default
  // child) whenever the product changes.
  useEffect(() => {
    if (!product) return;

    const axes = product.variationAxes || [];
    const variants = product.variants || [];
    const variantParam = searchParams.get("variant");
    const colorParam = searchParams.get("color");

    let initialVariant = variants.find((v: ProductVariant) => v.isDefault) || variants[0];
    if (variantParam) {
      const match = variants.find((v: ProductVariant) => v.id === variantParam || v.sku === variantParam);
      if (match) initialVariant = match;
    }

    const next: Record<string, string> = {};
    for (const axis of axes) {
      const fromVariant = initialVariant?.attributes?.[axis.key];
      let value = fromVariant !== undefined && fromVariant !== null ? String(fromVariant) : "";
      if (!value && axis.key === "color" && colorParam) {
        value = axis.values.find((v: string) => v.toLowerCase() === colorParam.toLowerCase()) || "";
      }
      if (!value) value = axis.values[0] || "";
      next[axis.key] = value;
    }
    setSelectedValues(next);
    setActiveImage(0);
  }, [product?.id, product?.slug, searchParams]);

  useEffect(() => {
    window.scrollTo(0, 0);
  }, [id]);

  useEffect(() => {
    if (!product?.id) return;
    if (user?.id) {
      recordProductView(String(product.id));
    } else {
      recordRecentlyViewed(String(product.slug || product.id));
    }
  }, [product?.id, product?.slug, user?.id]);

  useEffect(() => {
    let isMounted = true;
    const targetProductId = product?.id || (Array.isArray(id) ? id[0] : id);
    if (!targetProductId) return;

    catalogService
      .getProductReviews(String(targetProductId))
      .then((publicRes) => {
        if (!isMounted) return;
        const publicList = publicRes?.data ?? [];

        const mapped: ProductReview[] = publicList.map((r: Review) => {
          const isOwn = user?.id && String(r.userId) === String(user.id);
          return {
            id: r.id,
            author: isOwn ? (user.name || tProducts("details.you")) : r.user?.name || tProducts("details.customer"),
            avatar: isOwn ? null : r.user?.avatar || null,
            rating: Number(r.rating) || 0,
            date: r.createdAt
              ? new Date(r.createdAt).toLocaleDateString(locale, {
                  year: "numeric",
                  month: "long",
                  day: "numeric",
                })
              : tProducts("details.recently"),
            title: r.title || tProducts("details.customerReview"),
            content: r.body || "",
            helpful: r.helpfulCount ?? 0,
            verified: Boolean(r.orderItemId),
            userVoted: r.userVoted ?? false,
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
      .then((res) => {
        if (!isMounted) return;
        const eligible = Boolean(res.data?.canReview);
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
      .filter((p) => Boolean(product.category) && p.category === product.category && p.id !== product.id)
      .slice(0, 5);
  }, [product, allProducts]);

  if ((allLoading && !initialProduct) || productLoading) {
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
  const activeVariant = resolveVariantByAttributes(product, selectedValues);
  const { availableStock, isVariantInStock } = resolveVariantStock(product, activeVariant);
  const imageEntries = resolveProductImageEntries(product, activeVariant);
  const images = imageEntries.map((image) => image.url);
  const highlights = Array.isArray(product.highlights) ? product.highlights : [];
  const reviewCount = reviews.length || product.reviews || 0;
  const averageRating =
    reviews.length > 0
      ? reviews.reduce((total, review) => total + (Number(review.rating) || 0), 0) /
        reviews.length
      : Number(product.stars) || 0;

  const isWishlisted = Boolean(
    activeVariant?.id
      ? checkIsWishlisted(String(product?.id), activeVariant.id) || checkIsWishlisted(String(product?.slug), activeVariant.id)
      : (product?.id && checkIsWishlisted(String(product.id))) || (product?.slug && checkIsWishlisted(String(product.slug)))
  );

  // Selecting any axis value may change the child variant's image.
  const handleSelectAxis = (key: string, value: string) => {
    setSelectedValues((prev) => ({ ...prev, [key]: value }));
    setActiveImage(0);
  };

  // Selecting a gallery image also selects the variant associated with it.
  // Generic product images (not used by any variant) leave options alone.
  const handleSelectImage = (index: number) => {
    const entry = imageEntries[index];
    if (!entry?.attributes || !entry.variantId) {
      setActiveImage(index);
      return;
    }

    const nextValues = { ...selectedValues };
    for (const axis of product.variationAxes || []) {
      const value = entry.attributes[axis.key];
      if (value !== undefined && value !== null && value !== "") {
        nextValues[axis.key] = String(value);
      }
    }

    const nextVariant = resolveVariantByAttributes(product, nextValues);
    const nextImages = resolveProductImageEntries(product, nextVariant);
    const nextIndex = nextImages.findIndex((image) => image.url === entry.url);

    setSelectedValues(nextValues);
    setActiveImage(nextIndex >= 0 ? nextIndex : 0);
  };

  const resolvePrice = (val: string | number | undefined): number => {
    if (typeof val === "number" && !isNaN(val)) return val;
    if (!val) return 0;
    const parsed = parseFloat(String(val).replace(/[^0-9.-]+/g, ""));
    return isNaN(parsed) ? 0 : parsed;
  };

  const activeVariantPrice = activeVariant?.extractedPrice ?? activeVariant?.price;
  const numVariantPrice = resolvePrice(activeVariantPrice);
  const numProductPrice = resolvePrice(product.extractedPrice ?? product.price);
  const currentPrice = numVariantPrice > 0 ? numVariantPrice : numProductPrice;

  const hasVariantOldPrice =
    activeVariant?.extractedOriginalPrice !== undefined ||
    activeVariant?.oldPrice !== undefined ||
    activeVariant?.originalPrice !== undefined;
  const numVariantOldPrice = hasVariantOldPrice
    ? resolvePrice(
        activeVariant?.extractedOriginalPrice ??
        activeVariant?.oldPrice ??
        activeVariant?.originalPrice
      )
    : 0;
  const numProductOldPrice = resolvePrice(product.extractedOriginalPrice ?? product.oldPrice ?? product.originalPrice);

  // If an active variant is selected, use its old price (do not bleed another variant's price onto it).
  // Only fall back to product-level old price if no active variant is selected.
  const rawOldPrice = activeVariant
    ? (numVariantOldPrice > 0 ? numVariantOldPrice : undefined)
    : (numProductOldPrice > 0 ? numProductOldPrice : undefined);

  // Render old price whenever a valid positive old price is configured and differs from current price
  const currentOldPrice = rawOldPrice && rawOldPrice > 0 && rawOldPrice !== currentPrice ? rawOldPrice : undefined;

  const calculatedDiscount = currentOldPrice && currentOldPrice > currentPrice
    ? `${Math.round(((currentOldPrice - currentPrice) / currentOldPrice) * 100)}%`
    : undefined;
  const discountPercentage = activeVariant ? calculatedDiscount : (product.discountPercentage || calculatedDiscount);
  const selectedColorLabel = String(
    activeVariant?.attributes?.color ?? selectedValues.color ?? Object.values(selectedValues)[0] ?? "Default"
  );

  const handleAddToCart = async () => {
    if (!isVariantInStock) return;
    setIsAdding(true);
    const resolvedVariantId =
      activeVariant?.id ||
      (product.variants && product.variants.find((v: ProductVariant) => v.isDefault)?.id) ||
      (product.variants && product.variants[0]?.id) ||
      product.variantId ||
      product.defaultVariantId;
    const payload = {
      id: String(product.id),
      slug: product.slug,
      variantId: resolvedVariantId,
      productName: product.product,
      productPrice: formatMoney(currentPrice, currency),
      img: images[0],
      quantity,
      color: selectedColorLabel,
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
      (product.variants && product.variants.find((v: ProductVariant) => v.isDefault)?.id) ||
      (product.variants && product.variants[0]?.id) ||
      product.variantId ||
      product.defaultVariantId;
    const payload = {
      id: String(product.id),
      slug: product.slug,
      variantId: resolvedVariantId,
      productName: product.product,
      productPrice: formatMoney(currentPrice, currency),
      img: images[0],
      quantity,
      color: selectedColorLabel,
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

  const handleShare = async () => {
    const copied = await copyTextToClipboard(window.location.href);
    if (copied) {
      notify.success(tProducts("details.linkCopied"));
    } else {
      // Fall back to showing the URL so the user can copy it manually.
      notify.info(tProducts("details.shareUrl", { url: window.location.href }));
    }
  };

  const handleAddReview = async (newReviewData: { rating: number; title: string; content: string }) => {
    if (!product?.id) {
      notify.error(tReviews("missingProduct"));
      return false;
    }

    try {
      const res = await catalogService.submitReview({
        productId: String(product.id),
        rating: newReviewData.rating,
        title: newReviewData.title,
        body: newReviewData.content,
      });

      const serverReview = (res?.data ?? res) as {
        id?: string | number;
        verifiedPurchase?: boolean;
      };
      notify.success(tReviews("submitSuccess"));

      const newRev: ProductReview = {
        id: serverReview?.id || Date.now(),
        author: tProducts("details.you"),
        avatar: null,
        rating: newReviewData.rating,
        title: newReviewData.title,
        content: newReviewData.content,
        date: tProducts("details.reviewedOn", {
          date: new Date().toLocaleDateString(locale, {
            year: "numeric",
            month: "long",
            day: "numeric",
          }),
        }),
        helpful: 0,
        verified: Boolean(serverReview?.verifiedPurchase ?? false),
      };

      setReviews((prev) => [newRev, ...prev]);
      setCanReview(false);
      return true;
    } catch (err) {
      const e = err as { response?: { data?: { code?: string }; status?: number } } | undefined;
      const errCode = e?.response?.data?.code;
      const statusCode = e?.response?.status;

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

  const handleVoteHelpful = async (reviewId: string | number) => {
    if (!user) {
      notify.error(tReviews("errors.unauthorized"));
      return;
    }

    // Optimistic toggle
    setReviews((prev) =>
      prev.map((r) => {
        if (r.id === reviewId) {
          const isVoted = r.userVoted;
          return {
            ...r,
            helpful: isVoted ? Math.max(0, r.helpful - 1) : r.helpful + 1,
            userVoted: !isVoted,
          };
        }
        return r;
      })
    );

    try {
      const res = await catalogService.voteReviewHelpful(String(reviewId));
      if (res?.data) {
        setReviews((prev) =>
          prev.map((r) =>
            r.id === reviewId
              ? { ...r, helpful: res.data.helpfulCount, userVoted: res.data.voted }
              : r
          )
        );
      }
    } catch (err) {
      // Revert optimistic update on error
      setReviews((prev) =>
        prev.map((r) => {
          if (r.id === reviewId) {
            const isVoted = r.userVoted;
            return {
              ...r,
              helpful: isVoted ? Math.max(0, r.helpful - 1) : r.helpful + 1,
              userVoted: !isVoted,
            };
          }
          return r;
        })
      );
      notify.error(apiErrorMessage(err, "Failed to submit vote"));
    }
  };

  return (
    <AnimatedSection className="mx-auto w-full max-w-7xl">
      <CatalogFeedLoader />
      {/* Breadcrumb */}
      <nav className="mb-4 flex items-center gap-1.5 overflow-x-auto whitespace-nowrap pb-1 text-xs text-muted-foreground scrollbar-none sm:mb-6 sm:text-sm">
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
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-12 lg:gap-8 xl:gap-10">
        <div className="self-start lg:col-span-7 lg:sticky lg:top-24">
          <ProductGallery
            images={images}
            productName={productName}
            activeImage={activeImage}
            onSelectImage={handleSelectImage}
            discountPercentage={discountPercentage}
          />
        </div>

        <div className="flex flex-col gap-5 lg:col-span-5">
          <div className="rounded-2xl border border-border bg-card p-5 shadow-sm sm:p-7">
            <ProductActions
              productCategory={product.category}
              productName={productName}
              stars={averageRating}
              reviewsCount={reviewCount}
              price={currentPrice}
              oldPrice={currentOldPrice}
              currency={currency}
              variationAxes={product.variationAxes}
              selectedValues={selectedValues}
              onSelectAxis={handleSelectAxis}
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

          <ProductHighlights highlights={highlights} />
        </div>
      </div>

      {/* Specifications & Shipping Sections */}
      <ProductSpecs
        productName={productName}
        category={product.category}
        brand={typeof product.brand === "string" ? product.brand : product.brand?.name}
        description={product.description}
        attributes={activeVariant?.attributes}
        condition={product.condition}
        warranty={product.warranty}
      />

      {/* Reviews Section */}
      <div className="mt-10 sm:mt-14">
        <ProductReviews
          stars={averageRating}
          reviewsCount={reviewCount}
          reviews={reviews}
          canReview={canReview}
          onAddReview={handleAddReview}
          onVoteHelpful={handleVoteHelpful}
        />
      </div>

      {/* Similar Products */}
      {similarProducts.length > 0 && (
        <div className="mt-10 mb-6 sm:mt-14 sm:mb-8">
          <SectionTitle
            title={tProducts("details.similarProducts")}
            link={`/products?category=${product.categoryId || product.category}`}
          />
          <ProductsGrid
            products={similarProducts as unknown as CardDetails[]}
            className="grid-cols-2 sm:grid-cols-3 lg:grid-cols-5"
          />
        </div>
      )}
    </AnimatedSection>
  );
}
