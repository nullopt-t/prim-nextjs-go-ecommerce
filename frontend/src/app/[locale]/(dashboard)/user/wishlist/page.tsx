import { WishlistContent } from "@/features/wishlist";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Dashboard Wishlist | PRIM",
  description: "Your saved products",
};

export default function UserWishlistPage() {
  return <WishlistContent />;
}
