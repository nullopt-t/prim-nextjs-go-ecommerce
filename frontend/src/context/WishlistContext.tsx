"use client";

import { createContext, useContext, useState, useEffect, useCallback, useMemo } from "react";
import { wishlistService, WishlistItem } from "@/services/wishlist";
import { useAuthContext } from "@/context/AuthContext";
import { toast } from "sonner";

interface WishlistContextType {
  wishlistItems: WishlistItem[];
  wishlistCount: number;
  loading: boolean;
  isWishlisted: (productId: string, variantId?: string) => boolean;
  toggleWishlist: (productId: string, variantId?: string, productName?: string) => Promise<boolean>;
  addToWishlist: (productId: string, variantId?: string, productName?: string) => Promise<boolean>;
  removeFromWishlist: (productId: string, variantId?: string, productName?: string) => Promise<boolean>;
  removeWishlistItem: (itemId: string, productName?: string) => Promise<boolean>;
  clearWishlist: () => Promise<boolean>;
  refreshWishlist: () => Promise<void>;
}

const WishlistContext = createContext<WishlistContextType | null>(null);

export function WishlistProvider({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuthContext();
  const [wishlistItems, setWishlistItems] = useState<WishlistItem[]>([]);
  const [wishlistCount, setWishlistCount] = useState<number>(0);
  const [loading, setLoading] = useState(false);

  const productIdsSet = useMemo(() => {
    const set = new Set<string>();
    wishlistItems.forEach((item) => {
      if (item.productId) set.add(String(item.productId));
      if (item.product?.id) set.add(String(item.product.id));
      if (item.product?.slug) set.add(String(item.product.slug));
      if (item.variantId) {
        set.add(`${item.productId}:${item.variantId}`);
        if (item.product?.id) set.add(`${item.product.id}:${item.variantId}`);
        if (item.product?.slug) set.add(`${item.product.slug}:${item.variantId}`);
      }
    });
    return set;
  }, [wishlistItems]);

  const refreshWishlist = useCallback(async () => {
    if (!isAuthenticated) {
      setWishlistItems([]);
      setWishlistCount(0);
      return;
    }

    setLoading(true);
    try {
      const res = await wishlistService.getItems(1, 100);
      const items = res?.data || [];
      setWishlistItems(items);
      setWishlistCount(res?.meta?.totalItems ?? items.length);
    } catch {
      // In case user is not authenticated or network error
      setWishlistItems([]);
      setWishlistCount(0);
    } finally {
      setLoading(false);
    }
  }, [isAuthenticated]);

  useEffect(() => {
    refreshWishlist();
  }, [refreshWishlist]);

  const isWishlisted = useCallback(
    (productId: string, variantId?: string) => {
      if (!productId) return false;
      if (variantId) {
        return productIdsSet.has(`${productId}:${variantId}`);
      }
      return productIdsSet.has(String(productId));
    },
    [productIdsSet]
  );

  const addToWishlist = async (productId: string, variantId?: string, productName?: string): Promise<boolean> => {
    if (!isAuthenticated) {
      toast.error("Please sign in to save items to your wishlist");
      return false;
    }

    try {
      const res = await wishlistService.addItem(productId, variantId);
      if (res?.data) {
        setWishlistItems((prev) => [res.data, ...prev]);
        setWishlistCount((c) => c + 1);
      } else {
        await refreshWishlist();
      }
      toast.success(productName ? `Saved "${productName}" to wishlist!` : "Saved to wishlist!");
      return true;
    } catch (err: any) {
      toast.error(err?.message || "Failed to add to wishlist");
      return false;
    }
  };

  const removeFromWishlist = async (productId: string, variantId?: string, productName?: string): Promise<boolean> => {
    if (!isAuthenticated) return false;

    // Optimistic update
    setWishlistItems((prev) =>
      prev.filter((i) => {
        const matchesProduct = String(i.productId) === String(productId) || String(i.product?.id) === String(productId) || String(i.product?.slug) === String(productId);
        if (variantId) {
          return !(matchesProduct && String(i.variantId) === String(variantId));
        }
        return !matchesProduct;
      })
    );
    setWishlistCount((c) => Math.max(0, c - 1));

    try {
      await wishlistService.removeByProductId(productId, variantId);
      toast.info(productName ? `Removed "${productName}" from wishlist` : "Removed from wishlist");
      return true;
    } catch (err: any) {
      toast.error(err?.message || "Failed to remove from wishlist");
      await refreshWishlist();
      return false;
    }
  };

  const removeWishlistItem = async (itemId: string, productName?: string): Promise<boolean> => {
    if (!isAuthenticated) return false;

    // Optimistic update
    setWishlistItems((prev) => prev.filter((i) => String(i.id) !== String(itemId)));
    setWishlistCount((c) => Math.max(0, c - 1));

    try {
      await wishlistService.removeItem(itemId);
      toast.info(productName ? `Removed "${productName}" from wishlist` : "Removed from wishlist");
      return true;
    } catch (err: any) {
      toast.error(err?.message || "Failed to remove item");
      await refreshWishlist();
      return false;
    }
  };

  const toggleWishlist = async (productId: string, variantId?: string, productName?: string): Promise<boolean> => {
    if (isWishlisted(productId, variantId)) {
      return removeFromWishlist(productId, variantId, productName);
    } else {
      return addToWishlist(productId, variantId, productName);
    }
  };

  const clearWishlist = async (): Promise<boolean> => {
    if (!isAuthenticated) return false;

    setWishlistItems([]);
    setWishlistCount(0);

    try {
      await wishlistService.clearWishlist();
      toast.success("Wishlist cleared successfully");
      return true;
    } catch (err: any) {
      toast.error(err?.message || "Failed to clear wishlist");
      await refreshWishlist();
      return false;
    }
  };

  return (
    <WishlistContext.Provider
      value={{
        wishlistItems,
        wishlistCount,
        loading,
        isWishlisted,
        toggleWishlist,
        addToWishlist,
        removeFromWishlist,
        removeWishlistItem,
        clearWishlist,
        refreshWishlist,
      }}
    >
      {children}
    </WishlistContext.Provider>
  );
}

export function useWishlist() {
  const context = useContext(WishlistContext);
  if (!context) {
    throw new Error("useWishlist must be used within a WishlistProvider");
  }
  return context;
}
