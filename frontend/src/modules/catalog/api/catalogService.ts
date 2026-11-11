import { api } from "@/shared/api/client";
import type { ApiResponse, ApiPaginatedResponse } from "@/shared/api/types";
import type { CursorMeta } from "@/shared/lib/pagination";
import { buildProductQuery } from "@/shared/lib/productQuery";
import { buildReviewQuery, type ReviewQuery } from "@/shared/lib/reviewQuery";
import type {
  AdminCategory,
  Brand,
  Category,
  CategoryAttribute,
  Product,
  ProductDetails,
  ProductRatingSummary,
  ProductVariant,
  ProductVariantMedia,
  Review,
  ReviewEligibility,
  Tag,
} from "@/modules/catalog/types";

export interface GetProductsParams {
  pageSize?: number;
  page?: number;
  cursor?: string;
  category?: string;
  brand?: string;
  search?: string;
  inStock?: boolean;
  status?: string;
  minPrice?: number;
  maxPrice?: number;
  sort?: string;
  /** Selected variant attribute facets, e.g. { color: "black", size: "XL" }. */
  attributes?: Record<string, string>;
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
      return api.get<ApiPaginatedResponse<Product>>(`/api/v1/products${params ? `?${params}` : ""}`);
    }

    const qs = buildProductQuery(params ?? {}).toString();
    return api.get<ApiPaginatedResponse<Product>>(`/api/v1/products${qs ? `?${qs}` : ""}`);
  },

  getProductBySlug: (slug: string) =>
    api.get<ApiResponse<ProductDetails>>(`/api/v1/products/${slug}`),

  getCategoryAttributes: (categoryId: string) =>
    api.get<ApiResponse<CategoryAttribute[]>>(`/api/v1/categories/${encodeURIComponent(categoryId)}/attributes`),

  getCategories: (params?: CatalogListParams) =>
    api.get<ApiResponse<Category[]> | ApiPaginatedResponse<Category>>(
      `/api/v1/categories${buildCatalogListQuery(params ?? 50)}`
    ),

  // Admin listing includes categories with no products and taxonomy tree fields
  getAdminCategories: (params?: CatalogListParams) =>
    api.get<ApiResponse<AdminCategory[]> | ApiPaginatedResponse<AdminCategory>>(
      `/api/v1/admin/categories${buildCatalogListQuery(params ?? 100)}`
    ),

  getBrands: (params?: CatalogListParams) =>
    api.get<ApiResponse<Brand[]> | ApiPaginatedResponse<Brand>>(
      `/api/v1/brands${buildCatalogListQuery(params ?? 50)}`
    ),

  getTags: () => api.get<ApiResponse<Tag[]>>("/api/v1/tags"),

  getVariantById: (id: string) =>
    api.get<ApiResponse<ProductVariant>>(`/api/v1/variants/${id}`),

  getProductReviews: (productId: string, params?: ReviewQuery) => {
    const query = buildReviewQuery(params);
    const search = query.toString();
    return api.get<ApiPaginatedResponse<Review, CursorMeta>>(
      `/api/v1/reviews?productId=${encodeURIComponent(productId)}${search ? `&${search}` : ""}`,
    );
  },

  getProductRatingSummary: (slug: string) =>
    api.get<ApiResponse<ProductRatingSummary>>(`/api/v1/products/${slug}/reviews/summary`),

  submitReview: (payload: { productId: string; rating: number; title?: string; body?: string }) =>
    api.post<ApiResponse<Review>, { productId: string; rating: number; title?: string; body?: string }>(
      "/api/v1/reviews",
      payload
    ),

  checkReviewEligibility: (productId: string) =>
    api.get<ApiResponse<ReviewEligibility>>(`/api/v1/reviews/eligibility?productId=${productId}`),

  voteReviewHelpful: (reviewId: string) =>
    api.post<ApiResponse<{ voted: boolean; helpfulCount: number }>>(`/api/v1/reviews/${reviewId}/helpful`),

  createProduct: (payload: Record<string, unknown>) =>
    api.post<ApiResponse<ProductDetails>>("/api/v1/admin/products", payload),

  getAdminProducts: (params?: GetProductsParams | string) => {
    if (typeof params === "string") {
      return api.get<ApiPaginatedResponse<Product>>(`/api/v1/admin/products${params ? `?${params}` : ""}`);
    }

    const query = new URLSearchParams();
    if (params?.pageSize) query.append("pageSize", String(params.pageSize));
    if (params?.page) query.append("page", String(params.page));
    if (params?.cursor) query.append("cursor", params.cursor);
    if (params?.category) query.append("category", params.category);
    if (params?.brand) query.append("brand", params.brand);
    if (params?.search) query.append("search", params.search);
    if (params?.status) query.append("status", params.status);

    const qs = query.toString();
    return api.get<ApiPaginatedResponse<Product>>(`/api/v1/admin/products${qs ? `?${qs}` : ""}`);
  },

  getAdminProduct: (id: string) =>
    api.get<ApiResponse<ProductDetails>>(`/api/v1/admin/products/${id}`),

  updateProduct: (id: string, payload: Record<string, unknown>) =>
    api.patch<ApiResponse<ProductDetails>>(`/api/v1/admin/products/${id}`, payload),

  publishProduct: (id: string) =>
    api.post<ApiResponse<ProductDetails>>(`/api/v1/admin/products/${id}/publish`),

  archiveProduct: (id: string) =>
    api.post<ApiResponse<ProductDetails>>(`/api/v1/admin/products/${id}/archive`),

  deleteProduct: (id: string) => api.delete<void>(`/api/v1/admin/products/${id}`),

  // Product Media & Thumbnails
  uploadProductThumbnail: (productId: string, file: File) => {
    const fd = new FormData();
    fd.append("file", file);
    return api.upload<ApiResponse<{ url: string }>>(`/api/v1/admin/products/${productId}/thumbnail`, fd);
  },

  deleteProductThumbnail: (productId: string) =>
    api.delete<void>(`/api/v1/admin/products/${productId}/thumbnail`),

  // Product Variants
  createVariant: (productId: string, payload: Record<string, unknown>) =>
    api.post<ApiResponse<ProductVariant>>(`/api/v1/admin/products/${productId}/variants`, payload),

  createVariantsBatch: (productId: string, payload: Record<string, unknown>) =>
    api.post<ApiResponse<ProductVariant[]>>(`/api/v1/admin/products/${productId}/variants/batch`, payload),

  updateVariant: (variantId: string, payload: Record<string, unknown>) =>
    api.patch<ApiResponse<ProductVariant>>(`/api/v1/admin/variants/${variantId}`, payload),

  deleteVariant: (variantId: string) =>
    api.delete<void>(`/api/v1/admin/variants/${variantId}`),

  setDefaultVariant: (productId: string, variantId: string) =>
    api.post<ApiResponse<ProductVariant>>(`/api/v1/admin/products/${productId}/variants/default`, { variantId }),

  uploadVariantThumbnail: (variantId: string, file: File) => {
    const fd = new FormData();
    fd.append("file", file);
    return api.upload<ApiResponse<{ url: string }>>(`/api/v1/admin/variants/${variantId}/thumbnail`, fd);
  },

  deleteVariantThumbnail: (variantId: string) =>
    api.delete<void>(`/api/v1/admin/variants/${variantId}/thumbnail`),

  uploadVariantMedia: (variantId: string, file: File) => {
    const fd = new FormData();
    fd.append("file", file);
    return api.upload<ApiResponse<ProductVariantMedia>>(`/api/v1/admin/variants/${variantId}/media/upload`, fd);
  },

  detachVariantMedia: (variantId: string, mediaId: string) =>
    api.delete<void>(`/api/v1/admin/variants/${variantId}/media/${mediaId}`),

  getVariantMedia: (variantId: string) =>
    api.get<ApiResponse<ProductVariantMedia[]>>(`/api/v1/variants/${variantId}/media`),

  // Category Admin
  createCategory: (payload: { name: string; parentId?: string }) =>
    api.post<ApiResponse<AdminCategory>>("/api/v1/admin/categories", payload),

  updateCategory: (id: string, payload: { name?: string; parentId?: string }) =>
    api.patch<ApiResponse<AdminCategory>>(`/api/v1/admin/categories/${id}`, payload),

  deleteCategory: (id: string) => api.delete<void>(`/api/v1/admin/categories/${id}`),

  // Brand Admin
  createBrand: (payload: { name: string; link?: string }) =>
    api.post<ApiResponse<Brand>>("/api/v1/admin/brands", payload),

  updateBrand: (id: string, payload: { name?: string; link?: string }) =>
    api.patch<ApiResponse<Brand>>(`/api/v1/admin/brands/${id}`, payload),

  deleteBrand: (id: string) =>
    api.delete<void>(`/api/v1/admin/brands/${id}`),
};
