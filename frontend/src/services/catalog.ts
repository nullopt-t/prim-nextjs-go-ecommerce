import { api } from "@/api/client";

export interface GetProductsParams {
  pageSize?: number;
  page?: number;
  category?: string;
  brand?: string;
  search?: string;
  inStock?: boolean;
}

export const catalogService = {
  getProducts: (params?: GetProductsParams | string) => {
    if (typeof params === "string") {
      return api.get(`/api/v1/products${params ? `?${params}` : ""}`);
    }

    const query = new URLSearchParams();
    if (params?.pageSize) query.append("pageSize", String(params.pageSize));
    if (params?.page) query.append("page", String(params.page));
    if (params?.category) query.append("category", params.category);
    if (params?.brand) query.append("brand", params.brand);
    if (params?.search) query.append("search", params.search);
    if (params?.inStock !== undefined) query.append("inStock", String(params.inStock));

    const qs = query.toString();
    return api.get(`/api/v1/products${qs ? `?${qs}` : ""}`);
  },

  getProductBySlug: (slug: string) => api.get(`/api/v1/products/${slug}`),
  getCategories: (pageSize: number = 50) => api.get(`/api/v1/categories?pageSize=${pageSize}`),
  getBrands: (pageSize: number = 50) => api.get(`/api/v1/brands?pageSize=${pageSize}`),
  getTags: () => api.get("/api/v1/tags"),
  getVariantById: (id: string) => api.get(`/api/v1/variants/${id}`),
  getProductReviews: (slug: string, query?: string) =>
    api.get(`/api/v1/products/${slug}/reviews${query ? `?${query}` : ""}`),
  getProductRatingSummary: (slug: string) => api.get(`/api/v1/products/${slug}/reviews/summary`),
  submitReview: (payload: { productId: string; rating: number; title?: string; body?: string }) =>
    api.post("/api/v1/reviews", payload),
  createProduct: (payload: any) => api.post("/api/v1/products", payload),
  updateProduct: (id: string, payload: any) => api.patch(`/api/v1/products/${id}`, payload),
  deleteProduct: (id: string) => api.delete(`/api/v1/products/${id}`),
};
