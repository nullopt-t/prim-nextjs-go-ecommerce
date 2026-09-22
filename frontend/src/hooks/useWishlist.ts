"use client";

import useFetch from "./useFetch";
import { api } from "@/api/client";

export function useWishlist() {
  const { data, loading, errorMsg } = useFetch(() => api.get("/api/v1/wishlist"), "");
  return { wishlistItems: data, loading, errorMsg };
}
