import { Metadata } from "next";
import OrderDetailsContent from "@/features/user/components/ui/orderDetailsContent";

export const metadata: Metadata = {
  title: "Order Details | PRIM",
  description: "View detailed invoice, shipping, and items for your order",
};

export default function OrderDetailsPage() {
  return <OrderDetailsContent />;
}
