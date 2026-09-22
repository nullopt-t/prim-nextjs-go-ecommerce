import { WishlistContent } from "@/features/wishlist";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Your Wishlist | PRIM",
  description: "View and manage your saved items",
};

export default function WishlistPage() {
  return <WishlistContent />;
}
