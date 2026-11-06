import { api } from "@/shared/api/client";
import type { ApiMessageResponse, ApiPaginatedResponse, ApiResponse } from "@/shared/api/types";
import type {
  AdjustStockPayload,
  CreateSellerProductPayload,
  CreateSellerVariantPayload,
  RegisterSellerPayload,
  RequestPayoutPayload,
  SlugAvailability,
  UpdateSellerPayload,
  UpdateSellerProductPayload,
  UpdateSellerVariantPayload,
  Seller,
  SellerDashboard,
  SellerMediaUpload,
  SellerOrder,
  SellerOrderItem,
  SellerPayout,
  SellerProduct,
  SellerProductListParams,
  SellerProductResponse,
  SellerPromotion,
  CreateSellerPromotionPayload,
  SellerStock,
  SellerVariant,
} from "@/modules/seller/types";

function productQuery(params?: SellerProductListParams): string {
  const query = new URLSearchParams();
  if (params?.page) query.set("page", String(params.page));
  if (params?.pageSize) query.set("pageSize", String(params.pageSize));
  if (params?.status) query.set("status", params.status);
  if (params?.category) query.set("category", params.category);
  if (params?.brand) query.set("brand", params.brand);
  if (params?.cursor) query.set("cursor", params.cursor);
  const qs = query.toString();
  return qs ? `?${qs}` : "";
}

export const sellerService = {
  getProfile: () => api.get<ApiResponse<Seller>>("/api/v1/seller/profile"),
  checkSlugAvailability: (slug: string) =>
    api.get<ApiResponse<SlugAvailability>>(
      `/api/v1/seller/slug-availability?slug=${encodeURIComponent(slug)}`,
    ),
  register: (payload: RegisterSellerPayload) =>
    api.post<ApiResponse<Seller>, RegisterSellerPayload>("/api/v1/seller/register", payload),
  updateProfile: (payload: UpdateSellerPayload) =>
    api.patch<ApiResponse<Seller>, UpdateSellerPayload>("/api/v1/seller/profile", payload),

  getDashboard: () =>
    api.get<ApiResponse<SellerDashboard>>("/api/v1/seller/dashboard"),

  getOrders: (params?: { status?: string; page?: number; pageSize?: number }) => {
    const query = new URLSearchParams();
    if (params?.status) query.set("status", params.status);
    if (params?.page) query.set("page", String(params.page));
    if (params?.pageSize) query.set("pageSize", String(params.pageSize));
    const qs = query.toString();
    return api.get<ApiPaginatedResponse<SellerOrder>>(`/api/v1/seller/orders${qs ? `?${qs}` : ""}`);
  },
  getOrder: (id: string) =>
    api.get<ApiResponse<SellerOrder>>(`/api/v1/seller/orders/${id}`),

  getPayouts: () =>
    api.get<ApiResponse<SellerPayout[]>>("/api/v1/seller/payouts"),
  requestPayout: (payload: RequestPayoutPayload, idempotencyKey: string) =>
    api.post<ApiResponse<SellerPayout>, RequestPayoutPayload>("/api/v1/seller/payouts", payload, {
      headers: { "Idempotency-Key": idempotencyKey },
    }),

  getProducts: (params?: SellerProductListParams) =>
    api.get<ApiPaginatedResponse<SellerProduct>>(`/api/v1/seller/products${productQuery(params)}`),
  createProduct: (payload: CreateSellerProductPayload) =>
    api.post<ApiResponse<SellerProductResponse>, CreateSellerProductPayload>("/api/v1/seller/products", payload),
  updateProduct: (id: string, payload: UpdateSellerProductPayload) =>
    api.patch<ApiMessageResponse, UpdateSellerProductPayload>(`/api/v1/seller/products/${id}`, payload),
  deleteProduct: (id: string) =>
    api.delete<ApiMessageResponse>(`/api/v1/seller/products/${id}`),
  publishProduct: (id: string) =>
    api.post<ApiMessageResponse>(`/api/v1/seller/products/${id}/publish`),
  archiveProduct: (id: string) =>
    api.post<ApiMessageResponse>(`/api/v1/seller/products/${id}/archive`),

  getVariants: (productId: string, params?: { page?: number; pageSize?: number }) => {
    const query = new URLSearchParams();
    if (params?.page) query.set("page", String(params.page));
    if (params?.pageSize) query.set("pageSize", String(params.pageSize));
    const qs = query.toString();
    return api.get<ApiPaginatedResponse<SellerVariant>>(`/api/v1/seller/products/${productId}/variants${qs ? `?${qs}` : ""}`);
  },
  createVariant: (productId: string, payload: CreateSellerVariantPayload) =>
    api.post<ApiResponse<SellerVariant>, CreateSellerVariantPayload>(`/api/v1/seller/products/${productId}/variants`, payload),
  updateVariant: (variantId: string, payload: UpdateSellerVariantPayload) =>
    api.patch<ApiMessageResponse, UpdateSellerVariantPayload>(`/api/v1/seller/variants/${variantId}`, payload),
  deleteVariant: (variantId: string) =>
    api.delete<ApiMessageResponse>(`/api/v1/seller/variants/${variantId}`),
  adjustStock: (variantId: string, payload: AdjustStockPayload) =>
    api.post<ApiResponse<SellerStock>, AdjustStockPayload>(`/api/v1/seller/variants/${variantId}/stock`, payload),
  uploadMedia: (file: File) => {
    const formData = new FormData();
    formData.append("file", file);
    return api.upload<ApiResponse<SellerMediaUpload>>("/api/v1/seller/media", formData);
  },
  getPromotions: () => api.get<{ data: SellerPromotion[] }>("/api/v1/seller/promotions"),
  createPromotion: (payload: CreateSellerPromotionPayload) =>
    api.post<ApiResponse<SellerPromotion>, CreateSellerPromotionPayload>("/api/v1/seller/promotions", payload),
  setPromotionStatus: (id: string, status: "active" | "disabled" | "archived") =>
    api.patch<ApiMessageResponse, { status: string }>(`/api/v1/seller/promotions/${id}/status`, { status }),
  deletePromotion: (id: string) => api.delete<ApiMessageResponse>(`/api/v1/seller/promotions/${id}`),
};

export type {
  AdjustStockPayload,
  CreateSellerProductPayload,
  CreateSellerVariantPayload,
  RegisterSellerPayload,
  RequestPayoutPayload,
  SlugAvailability,
  UpdateSellerPayload,
  UpdateSellerProductPayload,
  UpdateSellerVariantPayload,
  Seller,
  SellerDashboard,
  SellerMediaUpload,
  SellerOrder,
  SellerOrderItem,
  SellerPayout,
  SellerProduct,
  SellerVariant,
  SellerStock,
};
