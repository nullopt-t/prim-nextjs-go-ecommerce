import { Metadata } from "next";
import { OrdersContent } from "@/features/user";

export const metadata: Metadata = {
  title: "My Orders | PRIM",
  description: "View and track your previous orders",
};

export default function OrdersPage() {
  return <OrdersContent />;
}
