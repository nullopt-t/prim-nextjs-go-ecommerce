import { api } from "@/api/client";

export interface GetProductsParams {
  pageSize?: number;
  page?: number;
  cursor?: string;
  category?: string;
  brand?: string;
  search?: string;
  inStock?: boolean;
}

type CatalogListParams = number | { pageSize?: number; cursor?: string; search?: string; keyset?: boolean };

function buildCatalogListQuery(params?: CatalogListParams): string {
  const query = new URLSearchParams();
  if (typeof params === "number") {
    query.set("pageSize", String(params));
  } else if (params) {
    if (params.cursor) query.set("cursor", params.cursor);
    if (params.pageSize) query.set("pageSize", String(params.pageSize));
    if (params.search) query.set("search", params.search);
    if (params.keyset) query.set("keyset", "true");
  }
  const qs = query.toString();
  return qs ? `?${qs}` : "";
}

export const catalogService = {
  getProducts: (params?: GetProductsParams | string) => {
    if (typeof params === "string") {
      return api.get(`/api/v1/products${params ? `?${params}` : ""}`);
    }

    const query = new URLSearchParams();
    if (params?.pageSize) query.append("pageSize", String(params.pageSize));
    if (params?.page) query.append("page", String(params.page));
    if (params?.cursor) query.append("cursor", params.cursor);
    if (params?.category) query.append("category", params.category);
    if (params?.brand) query.append("brand", params.brand);
    if (params?.search) query.append("search", params.search);
    if (params?.inStock !== undefined) query.append("inStock", String(params.inStock));

    const qs = query.toString();
    return api.get(`/api/v1/products${qs ? `?${qs}` : ""}`);
  },

  getProductBySlug: (slug: string) => api.get(`/api/v1/products/${slug}`),
  getCategories: (params?: CatalogListParams) => api.get(`/api/v1/categories${buildCatalogListQuery(params ?? 50)}`),
  getBrands: (params?: CatalogListParams) => api.get(`/api/v1/brands${buildCatalogListQuery(params ?? 50)}`),
  getTags: () => api.get("/api/v1/tags"),
  getVariantById: (id: string) => api.get(`/api/v1/variants/${id}`),
  getProductReviews: (slug: string, query?: string) =>
    api.get(`/api/v1/products/${slug}/reviews${query ? `?${query}` : ""}`),
  getProductRatingSummary: (slug: string) => api.get(`/api/v1/products/${slug}/reviews/summary`),
  submitReview: (payload: { productId: string; rating: number; title?: string; body?: string }) =>
    api.post("/api/v1/reviews", payload),
  checkReviewEligibility: (productId: string) =>
    api.get<{ canReview: boolean }>(`/api/v1/reviews/eligibility?productId=${productId}`),
  createProduct: (payload: any) => api.post("/api/v1/admin/products", payload),
  updateProduct: (id: string, payload: any) => api.patch(`/api/v1/admin/products/${id}`, payload),
  deleteProduct: (id: string) => api.delete(`/api/v1/admin/products/${id}`),

  // Category Admin
  createCategory: (payload: { name: string; parentId?: string }) =>
    api.post("/api/v1/admin/categories", payload),
  updateCategory: (id: string, payload: { name?: string; parentId?: string }) =>
    api.patch(`/api/v1/admin/categories/${id}`, payload),
  deleteCategory: (id: string) => api.delete(`/api/v1/admin/categories/${id}`),

  // Brand Admin
  createBrand: (payload: { name: string; link?: string }) =>
    api.post("/api/v1/admin/brands", payload),
  updateBrand: (id: string, payload: { name?: string; link?: string }) =>
    api.patch(`/api/v1/admin/brands/${id}`, payload),
  deleteBrand: (id: string) =>
    api.delete(`/api/v1/admin/brands/${id}`),
};
