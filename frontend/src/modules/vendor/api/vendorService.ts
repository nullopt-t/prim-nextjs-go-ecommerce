import { api } from "@/shared/api/client";
import type { ApiMessageResponse, ApiPaginatedResponse, ApiResponse } from "@/shared/api/types";
import type {
  AdjustStockPayload,
  CreateVendorProductPayload,
  CreateVendorVariantPayload,
  RegisterVendorPayload,
  RequestPayoutPayload,
  SlugAvailability,
  UpdateVendorPayload,
  UpdateVendorProductPayload,
  UpdateVendorVariantPayload,
  Vendor,
  VendorDashboard,
  VendorMediaUpload,
  VendorOrder,
  VendorOrderItem,
  VendorPayout,
  VendorProduct,
  VendorProductListParams,
  VendorProductResponse,
  VendorStock,
  VendorVariant,
} from "@/modules/vendor/types";

function productQuery(params?: VendorProductListParams): string {
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

export const vendorService = {
  getProfile: () => api.get<ApiResponse<Vendor>>("/api/v1/vendor/profile"),
  checkSlugAvailability: (slug: string) =>
    api.get<ApiResponse<SlugAvailability>>(
      `/api/v1/vendor/slug-availability?slug=${encodeURIComponent(slug)}`,
    ),
  register: (payload: RegisterVendorPayload) =>
    api.post<ApiResponse<Vendor>, RegisterVendorPayload>("/api/v1/vendor/register", payload),
  updateProfile: (payload: UpdateVendorPayload) =>
    api.patch<ApiResponse<Vendor>, UpdateVendorPayload>("/api/v1/vendor/profile", payload),

  getDashboard: () =>
    api.get<ApiResponse<VendorDashboard>>("/api/v1/vendor/dashboard"),

  getOrders: (params?: { status?: string; page?: number; pageSize?: number }) => {
    const query = new URLSearchParams();
    if (params?.status) query.set("status", params.status);
    if (params?.page) query.set("page", String(params.page));
    if (params?.pageSize) query.set("pageSize", String(params.pageSize));
    const qs = query.toString();
    return api.get<ApiPaginatedResponse<VendorOrder>>(`/api/v1/vendor/orders${qs ? `?${qs}` : ""}`);
  },
  getOrder: (id: string) =>
    api.get<ApiResponse<VendorOrder>>(`/api/v1/vendor/orders/${id}`),

  getPayouts: () =>
    api.get<ApiResponse<VendorPayout[]>>("/api/v1/vendor/payouts"),
  requestPayout: (payload: RequestPayoutPayload, idempotencyKey: string) =>
    api.post<ApiResponse<VendorPayout>, RequestPayoutPayload>("/api/v1/vendor/payouts", payload, {
      headers: { "Idempotency-Key": idempotencyKey },
    }),

  getProducts: (params?: VendorProductListParams) =>
    api.get<ApiPaginatedResponse<VendorProduct>>(`/api/v1/vendor/products${productQuery(params)}`),
  createProduct: (payload: CreateVendorProductPayload) =>
    api.post<ApiResponse<VendorProductResponse>, CreateVendorProductPayload>("/api/v1/vendor/products", payload),
  updateProduct: (id: string, payload: UpdateVendorProductPayload) =>
    api.patch<ApiMessageResponse, UpdateVendorProductPayload>(`/api/v1/vendor/products/${id}`, payload),
  deleteProduct: (id: string) =>
    api.delete<ApiMessageResponse>(`/api/v1/vendor/products/${id}`),
  publishProduct: (id: string) =>
    api.post<ApiMessageResponse>(`/api/v1/vendor/products/${id}/publish`),
  archiveProduct: (id: string) =>
    api.post<ApiMessageResponse>(`/api/v1/vendor/products/${id}/archive`),

  getVariants: (productId: string, params?: { page?: number; pageSize?: number }) => {
    const query = new URLSearchParams();
    if (params?.page) query.set("page", String(params.page));
    if (params?.pageSize) query.set("pageSize", String(params.pageSize));
    const qs = query.toString();
    return api.get<ApiPaginatedResponse<VendorVariant>>(`/api/v1/vendor/products/${productId}/variants${qs ? `?${qs}` : ""}`);
  },
  createVariant: (productId: string, payload: CreateVendorVariantPayload) =>
    api.post<ApiResponse<VendorVariant>, CreateVendorVariantPayload>(`/api/v1/vendor/products/${productId}/variants`, payload),
  updateVariant: (variantId: string, payload: UpdateVendorVariantPayload) =>
    api.patch<ApiMessageResponse, UpdateVendorVariantPayload>(`/api/v1/vendor/variants/${variantId}`, payload),
  deleteVariant: (variantId: string) =>
    api.delete<ApiMessageResponse>(`/api/v1/vendor/variants/${variantId}`),
  adjustStock: (variantId: string, payload: AdjustStockPayload) =>
    api.post<ApiResponse<VendorStock>, AdjustStockPayload>(`/api/v1/vendor/variants/${variantId}/stock`, payload),
  uploadMedia: (file: File) => {
    const formData = new FormData();
    formData.append("file", file);
    return api.upload<ApiResponse<VendorMediaUpload>>("/api/v1/vendor/media", formData);
  },
};

export type {
  AdjustStockPayload,
  CreateVendorProductPayload,
  CreateVendorVariantPayload,
  RegisterVendorPayload,
  RequestPayoutPayload,
  SlugAvailability,
  UpdateVendorPayload,
  UpdateVendorProductPayload,
  UpdateVendorVariantPayload,
  Vendor,
  VendorDashboard,
  VendorMediaUpload,
  VendorOrder,
  VendorOrderItem,
  VendorPayout,
  VendorProduct,
  VendorVariant,
  VendorStock,
};
