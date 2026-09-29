"use client";

import useFetch from "./useFetch";
import { wishlistService } from "@/services/wishlist";

export function useWishlist() {
  const { data, loading, errorMsg } = useFetch(() => wishlistService.getItems(), "");
  return { wishlistItems: data, loading, errorMsg };
}
