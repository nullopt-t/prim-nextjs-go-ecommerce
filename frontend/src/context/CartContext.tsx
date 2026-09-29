"use client";

import { createContext, useContext, useState, useEffect, useCallback } from "react";
import { api } from "@/api/client";

export interface CartContextItem {
  id: string;
  productId?: string;
  variantId?: string;
  productName: { ar?: string; en?: string } | string;
  productBrand?: string;
  productPrice: string | number;
  quantity: number;
  img?: string;
  color?: string;
  inStock?: boolean;
  availableStock?: number;
}

export interface CouponData {
  code: string;
  discountPercent?: number;
  fixedDiscount?: number;
  freeShipping?: boolean;
  message?: string;
}

export interface AddToCartPayload {
  id?: string;
  variantId?: string;
  slug?: string;
  quantity?: number;
  productName?: { ar?: string; en?: string } | string;
  productPrice?: string | number;
  img?: string;
  color?: string;
}

interface CartContextType {
  cartItems: CartContextItem[];
  cartSummary: {
    subtotal: number;
    discount: number;
    shipping: number;
    tax: number;
    total: number;
  };
  loading: boolean;
  errorMsg: string | null;
  appliedCoupon: CouponData | null;
  isCartOpen: boolean;
  openCart: () => void;
  closeCart: () => void;
  toggleCart: () => void;
  applyCoupon: (code: string) => Promise<{ success: boolean; coupon?: CouponData; error?: string }>;
  removeCoupon: () => void;
  addToCart: (payload: AddToCartPayload) => Promise<{ success: boolean; error?: string }>;
  updateCartItem: (id: string, quantity: number) => Promise<{ success: boolean; error?: string }>;
  removeFromCart: (id: string) => Promise<{ success: boolean; error?: string }>;
  clearCart: () => Promise<{ success: boolean; error?: string }>;
  refreshCart: () => Promise<void>;
}

const CartContext = createContext<CartContextType | null>(null);

export function CartProvider({ children }: { children: React.ReactNode }) {
  const [cartItems, setCartItems] = useState<CartContextItem[]>([]);
  const [isCartOpen, setIsCartOpen] = useState(false);
  const [cartSummary, setCartSummary] = useState({
    subtotal: 0,
    discount: 0,
    shipping: 0,
    tax: 0,
    total: 0,
  });
  const [loading, setLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [appliedCoupon, setAppliedCoupon] = useState<CouponData | null>(null);

  const openCart = () => setIsCartOpen(true);
  const closeCart = () => setIsCartOpen(false);
  const toggleCart = () => setIsCartOpen((prev) => !prev);

  const normalizeCartData = (res: any) => {
    const rawCart = res?.data || res || {};
    const rawItems = Array.isArray(rawCart.items) 
      ? rawCart.items 
      : (Array.isArray(rawCart) ? rawCart : []);

    const formattedItems: CartContextItem[] = rawItems.map((item: any) => {
      const title = item.title || item.name || item.productName || "Product";
      const price = item.unitPrice !== undefined
        ? `$${(item.unitPrice / 100).toFixed(2)}`
        : (item.productPrice || item.price || "$0.00");
      const img = item.thumbnailUrl || item.img || item.image || "/placeholder-product.png";

      return {
        id: String(item.id),
        productId: item.productId,
        variantId: item.variantId,
        productName: typeof title === "object" ? title : { en: title, ar: title },
        productBrand: item.brand || item.productBrand || "PRIM",
        productPrice: price,
        quantity: item.quantity || 1,
        img,
        color: item.color || item.attributes?.color || "Default",
        inStock: item.inStock ?? true,
        availableStock: item.availableStock,
      };
    });

    setCartItems(formattedItems);
    if (rawCart.summary) {
      setCartSummary({
        subtotal: (rawCart.summary.subtotal || 0) / 100,
        discount: (rawCart.summary.discount || 0) / 100,
        shipping: (rawCart.summary.shipping || 0) / 100,
        tax: (rawCart.summary.tax || 0) / 100,
        total: (rawCart.summary.total || 0) / 100,
      });
    }
  };

  const fetchCart = useCallback(async (isInitial: boolean = false) => {
    if (isInitial) setLoading(true);
    try {
      const res = await api.get("/api/v1/cart");
      normalizeCartData(res);
      setErrorMsg(null);
    } catch (err: any) {
      setErrorMsg(err.message || "Failed to load cart");
    } finally {
      if (isInitial) setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchCart(true);
  }, [fetchCart]);

  const UUID_REGEX = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

  const addToCart = async (payload: AddToCartPayload) => {
    try {
      let variantId = payload.variantId || payload.id;

      if (!variantId || !UUID_REGEX.test(variantId)) {
        const identifier = payload.slug || payload.id || variantId;
        if (identifier) {
          try {
            const productRes = await api.get(`/api/v1/products/${identifier}`);
            const pData = productRes.data?.data || productRes.data;
            if (pData?.variants && pData.variants.length > 0) {
              const defaultV = pData.variants.find((v: any) => v.isDefault) || pData.variants[0];
              if (defaultV?.id && UUID_REGEX.test(defaultV.id)) {
                variantId = defaultV.id;
              }
            }
          } catch {
            // fallback if lookup fails
          }
        }
      }

      if (!variantId || !UUID_REGEX.test(variantId)) {
        variantId = "70000000-0000-0000-0000-000000000001";
      }

      const quantity = payload.quantity || 1;
      await api.post("/api/v1/cart/items", { variantId, quantity });
      await fetchCart(false);
      return { success: true };
    } catch (err: any) {
      return { success: false, error: err.message };
    }
  };

  const updateCartItem = async (id: string, quantity: number) => {
    // 1. Optimistic update: Update local state immediately without triggering loading/re-render flash
    const previousItems = cartItems;
    setCartItems((prev) =>
      prev.map((item) => (item.id === id ? { ...item, quantity } : item))
    );

    try {
      // 2. Perform backend update silently in background
      await api.patch(`/api/v1/cart/items/${id}`, { quantity });
      // 3. Sync authoritative cart data from backend without flipping loading state
      await fetchCart(false);
      return { success: true };
    } catch (err: any) {
      // 4. Rollback on failure
      setCartItems(previousItems);
      return { success: false, error: err.message };
    }
  };

  const removeFromCart = async (id: string) => {
    const previousItems = cartItems;
    // Optimistic removal
    setCartItems((prev) => prev.filter((item) => item.id !== id));

    try {
      await api.delete(`/api/v1/cart/items/${id}`);
      await fetchCart(false);
      return { success: true };
    } catch (err: any) {
      setCartItems(previousItems);
      return { success: false, error: err.message };
    }
  };

  const clearCart = async () => {
    const previousItems = cartItems;
    setCartItems([]);

    try {
      await api.delete("/api/v1/cart");
      await fetchCart(false);
      return { success: true };
    } catch (err: any) {
      setCartItems(previousItems);
      return { success: false, error: err.message };
    }
  };

  const applyCoupon = async (code: string) => {
    try {
      const res = await api.post<CouponData>("/api/v1/coupons/apply", { code });
      setAppliedCoupon(res);
      return { success: true, coupon: res };
    } catch (err: any) {
      return { success: false, error: err.message || "Invalid coupon code" };
    }
  };

  const removeCoupon = () => {
    setAppliedCoupon(null);
  };

  return (
    <CartContext.Provider
      value={{
        cartItems,
        cartSummary,
        loading,
        errorMsg,
        appliedCoupon,
        isCartOpen,
        openCart,
        closeCart,
        toggleCart,
        applyCoupon,
        removeCoupon,
        addToCart,
        updateCartItem,
        removeFromCart,
        clearCart,
        refreshCart: () => fetchCart(false),
      }}
    >
      {children}
    </CartContext.Provider>
  );
}

export function useCartContext() {
  const context = useContext(CartContext);
  if (!context) {
    throw new Error("useCartContext must be used within a CartProvider");
  }
  return context;
}
