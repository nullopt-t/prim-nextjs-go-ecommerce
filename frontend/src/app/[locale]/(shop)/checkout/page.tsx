import { Checkout } from "@/features/checkout";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Secure Checkout | PRIM",
  description: "Complete your order with secure delivery and payment options",
};

export default function CheckoutPage() {
  return <Checkout />;
}
