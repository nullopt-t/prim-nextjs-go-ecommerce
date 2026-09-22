"use client";

import { useCartContext } from "@/context/CartContext";
import { useState } from "react";

export function useCart() {
  const {
    cartItems,
    cartSummary,
    loading,
    errorMsg,
    refreshCart,
    appliedCoupon,
    applyCoupon,
    removeCoupon,
    addToCart,
    updateCartItem,
    removeFromCart,
    clearCart,
    isCartOpen,
    openCart,
    closeCart,
    toggleCart,
  } = useCartContext();

  return {
    cartItems,
    cartSummary,
    loading,
    errorMsg,
    refreshCart,
    appliedCoupon,
    applyCoupon,
    removeCoupon,
    addToCart,
    updateCartItem,
    removeFromCart,
    clearCart,
    isCartOpen,
    openCart,
    closeCart,
    toggleCart,
  };
}

export function useAddToCart() {
  const { addToCart } = useCartContext();
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const mutate = async (payload: any) => {
    setLoading(true);
    setErrorMsg(null);
    const result = await addToCart(payload);
    if (!result.success) setErrorMsg(result.error || "Failed to add to cart");
    setLoading(false);
    return result;
  };

  return { addToCart: mutate, loading, errorMsg };
}

export function useUpdateCartItem() {
  const { updateCartItem } = useCartContext();
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const mutate = async ({ id, quantity }: { id: string; quantity: number }) => {
    setLoading(true);
    setErrorMsg(null);
    const result = await updateCartItem(id, quantity);
    if (!result.success) setErrorMsg(result.error || "Failed to update item");
    setLoading(false);
    return result;
  };

  return { updateCartItem: mutate, loading, errorMsg };
}

export function useRemoveFromCart() {
  const { removeFromCart } = useCartContext();
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const mutate = async (id: string) => {
    setLoading(true);
    setErrorMsg(null);
    const result = await removeFromCart(id);
    if (!result.success) setErrorMsg(result.error || "Failed to remove item");
    setLoading(false);
    return result;
  };

  return { removeFromCart: mutate, loading, errorMsg };
}

export function useClearCart() {
  const { clearCart } = useCartContext();
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const mutate = async () => {
    setLoading(true);
    setErrorMsg(null);
    const result = await clearCart();
    if (!result.success) setErrorMsg(result.error || "Failed to clear cart");
    setLoading(false);
    return result;
  };

  return { clearCart: mutate, loading, errorMsg };
}
