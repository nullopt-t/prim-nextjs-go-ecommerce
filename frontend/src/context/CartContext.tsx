"use client";

import { createContext, useContext, useState, useEffect, useCallback } from "react";
import { cartService } from "@/services/cart";
import { useAuthContext } from "@/context/AuthContext";

export interface CartContextItem {
  id: string;
  productId?: string;
  productSlug?: string;
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
  id?: string;
  code: string;
  discountType?: "percentage" | "fixed_amount" | string;
  discountValue?: number;
  discountPercent?: number;
  fixedDiscount?: number;
  minCartAmount?: number;
  discountAmount?: number;
  subtotal?: number;
  shipping?: number;
  total?: number;
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
  resetCart: () => void;
}

const CartContext = createContext<CartContextType | null>(null);

export function CartProvider({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, user, loading: authLoading } = useAuthContext();
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

  const resetCart = useCallback(() => {
    setCartItems([]);
    setCartSummary({
      subtotal: 0,
      discount: 0,
      shipping: 0,
      tax: 0,
      total: 0,
    });
    setAppliedCoupon(null);
  }, []);

  const normalizeCartData = (res: any) => {
    const rawCart = res?.data || res || {};
    const rawItems = Array.isArray(rawCart.items) 
      ? rawCart.items 
      : (Array.isArray(rawCart) ? rawCart : []);

    const formattedItems: CartContextItem[] = rawItems.map((item: any) => {
      const title = item.title || item.name || item.productName || "Product";
      const price = item.unitPrice !== undefined
        ? `EGP ${(item.unitPrice / 100).toFixed(2)}`
        : (item.productPrice || item.price || "EGP 0.00");
      const img = item.thumbnailUrl || item.img || item.image || "/placeholder-product.png";

      return {
        id: String(item.id),
        productId: item.productId,
        productSlug: item.productSlug || item.slug,
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
    } else {
      setCartSummary({
        subtotal: 0,
        discount: 0,
        shipping: 0,
        tax: 0,
        total: 0,
      });
    }
  };

  const fetchCart = useCallback(async (isInitial: boolean = false) => {
    if (isInitial) setLoading(true);
    try {
      const res = await cartService.getCart();
      normalizeCartData(res);
      setErrorMsg(null);
    } catch (err: any) {
      setErrorMsg(err.message || "Failed to load cart");
    } finally {
      if (isInitial) setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!authLoading) {
      fetchCart(true);
    }
  }, [authLoading, user, fetchCart]);

  const addToCart = async (payload: AddToCartPayload) => {
    try {
      await cartService.addItem({
        variantId: payload.variantId,
        id: payload.id,
        slug: payload.slug,
        quantity: payload.quantity || 1,
      });
      await fetchCart(false);
      return { success: true };
    } catch (err: any) {
      return { success: false, error: err.message };
    }
  };

  const updateCartItem = async (id: string, quantity: number) => {
    // 1. Optimistic update
    const previousItems = cartItems;
    setCartItems((prev) =>
      prev.map((item) => (item.id === id ? { ...item, quantity } : item))
    );

    try {
      // 2. Delegate directly to cartService
      await cartService.updateItemQuantity(id, quantity);
      // 3. Sync authoritative cart data from backend without flipping loading state
      await fetchCart(false);
      return { success: true };
    } catch (err: any) {
      // 4. If the item was not found on the server (stale cart / re-seeded session), sync authoritative state
      await fetchCart(false);
      return { success: false, error: err.message };
    }
  };

  const removeFromCart = async (id: string) => {
    const previousItems = cartItems;
    setCartItems((prev) => prev.filter((item) => item.id !== id));

    try {
      await cartService.removeItem(id);
      await fetchCart(false);
      return { success: true };
    } catch (err: any) {
      // If already deleted or stale, re-sync with backend
      await fetchCart(false);
      return { success: false, error: err.message };
    }
  };

  const clearCart = async () => {
    const previousItems = cartItems;
    setCartItems([]);

    try {
      await cartService.clearCart();
      await fetchCart(false);
      return { success: true };
    } catch (err: any) {
      setCartItems(previousItems);
      return { success: false, error: err.message };
    }
  };

  const applyCoupon = async (code: string) => {
    try {
      const currentSubtotal = cartSummary.subtotal || cartItems.reduce((acc, item) => {
        const raw = typeof item.productPrice === "number"
          ? item.productPrice
          : parseFloat(String(item.productPrice || "").replace(/[^0-9.-]+/g, "")) || 0;
        return acc + raw * (item.quantity || 1);
      }, 0);

      const res: any = await cartService.applyCoupon(code, currentSubtotal);
      const raw: any = res?.data || res || {};
      
      const discountType = raw.discountType || (raw.discountPercent ? "percentage" : "fixed_amount");
      const discountValue = Number(raw.discountValue || 0);
      const discountAmount = raw.discountAmount !== undefined ? Number(raw.discountAmount) / 100 : undefined;
      const serverSubtotal = raw.subtotal !== undefined ? Number(raw.subtotal) / 100 : undefined;
      const serverShipping = raw.shipping !== undefined ? Number(raw.shipping) / 100 : undefined;
      const serverTotal = raw.total !== undefined ? Number(raw.total) / 100 : undefined;

      const couponData: CouponData = {
        id: raw.id,
        code: raw.code || code.toUpperCase(),
        discountType,
        discountValue,
        discountPercent: raw.discountPercent ?? (discountType === "percentage" ? discountValue : undefined),
        fixedDiscount: raw.fixedDiscount ?? (discountType === "fixed_amount" ? discountValue / 100 : undefined),
        minCartAmount: raw.minCartAmount ? Number(raw.minCartAmount) / 100 : undefined,
        discountAmount,
        subtotal: serverSubtotal,
        shipping: serverShipping,
        total: serverTotal,
        freeShipping: Boolean(raw.freeShipping),
        message: raw.message,
      };

      setAppliedCoupon(couponData);

      // Server is the authoritative source for calculation: update cartSummary directly from server
      if (serverSubtotal !== undefined && discountAmount !== undefined && serverTotal !== undefined) {
        setCartSummary((prev) => ({
          ...prev,
          subtotal: serverSubtotal,
          discount: discountAmount,
          shipping: serverShipping ?? prev.shipping,
          total: serverTotal,
        }));
      }

      return { success: true, coupon: couponData };
    } catch (err: any) {
      return { success: false, error: err.message || "Invalid coupon code" };
    }
  };

  const removeCoupon = () => {
    setAppliedCoupon(null);
    setCartSummary((prev) => ({
      ...prev,
      discount: 0,
      total: prev.subtotal + prev.shipping + prev.tax,
    }));
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
        resetCart,
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
