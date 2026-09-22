import CartLayout from "@/features/cart/components/layout/cartLayout";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Your Shopping Cart | PRIM",
  description: "Review and checkout your cart items",
};

export default function CartPage() {
  return <CartLayout />;
}
