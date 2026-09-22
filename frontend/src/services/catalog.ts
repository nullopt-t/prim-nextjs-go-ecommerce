import { api } from "@/api/client";

export const catalogService = {
	getProducts: (params?: string) => api.get(`/api/v1/products${params ? `?${params}` : ""}`),
	getProductBySlug: (slug: string) => api.get(`/api/v1/products/${slug}`),
	getCategories: () => api.get("/api/v1/categories"),
	getBrands: () => api.get("/api/v1/brands"),
	getTags: () => api.get("/api/v1/tags"),
	getVariantById: (id: string) => api.get(`/api/v1/variants/${id}`),
	getProductReviews: (slug: string, query?: string) => api.get(`/api/v1/products/${slug}/reviews${query ? `?${query}` : ""}`),
	getProductRatingSummary: (slug: string) => api.get(`/api/v1/products/${slug}/reviews/summary`),
	submitReview: (payload: { productId: string; rating: number; title?: string; body?: string }) => api.post("/api/v1/reviews", payload),
	createProduct: (payload: any) => api.post("/api/v1/products", payload),
	updateProduct: (id: string, payload: any) => api.patch(`/api/v1/products/${id}`, payload),
	deleteProduct: (id: string) => api.delete(`/api/v1/products/${id}`),
};
