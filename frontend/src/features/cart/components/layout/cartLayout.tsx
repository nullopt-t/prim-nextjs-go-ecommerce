"use client";

import { useCart } from "@/hooks/useCart";
import CartTitle from "@/features/cart/components/ui/cartTitle";
import PaymentBox from "@/features/cart/components/ui/paymentBox";
import OrdersGrid from "@/features/cart/components/ui/ordersGrid";
import Coupon from "@/features/cart/components/ui/coupon";

export default function CartLayout() {
  const { cartItems, loading } = useCart();
  const hasItems = !loading && cartItems && cartItems.length > 0;

  return (
    <div className="max-w-5xl mx-auto flex flex-col gap-8 py-6">
      <CartTitle />
      {hasItems ? (
        <div className="flex flex-col lg:flex-row gap-8 items-start">
          <div className="w-full lg:w-2/3 flex flex-col gap-6">
            <OrdersGrid />
            <Coupon />
          </div>
          <div className="w-full lg:w-1/3 sticky top-24">
            <PaymentBox />
          </div>
        </div>
      ) : (
        <div className="w-full">
          <OrdersGrid />
        </div>
      )}
    </div>
  );
}
