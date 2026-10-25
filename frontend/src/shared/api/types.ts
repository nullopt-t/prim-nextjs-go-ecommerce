import type { CursorMeta } from "@/shared/lib/pagination";

export interface OffsetMeta {
  page: number;
  pageSize: number;
  totalItems: number;
  totalPages: number;
  hasPrevious?: boolean;
  hasNext?: boolean;
}

export type PaginationMeta = CursorMeta | OffsetMeta;

/**
 * Standard single-resource envelope returned by the backend.
 * Corresponds to Go's `api.DataResponse` / `api.SuccessResponse`.
 */
export interface ApiResponse<T> {
  data: T;
  message?: string;
  meta?: Record<string, unknown> | PaginationMeta;
}

/**
 * Standard paginated collection envelope returned by the backend.
 * Corresponds to Go's `api.PaginatedResponse`.
 */
export interface ApiPaginatedResponse<T, M = PaginationMeta> {
  data: T[];
  meta: M;
}

/**
 * Informational acknowledgment envelope.
 * Corresponds to Go's `api.MessageResponse`.
 */
export interface ApiMessageResponse {
  message: string;
}

/**
 * Field-level validation error detail.
 * Corresponds to Go's `api.FieldError`.
 */
export interface ApiFieldError {
  field?: string;
  constraint?: string;
  message?: string;
}

/**
 * Standard API error response envelope.
 * Corresponds to Go's `api.ErrorResponse` / `api.BadReqResponse`.
 */
export interface ApiErrorResponse {
  code?: string;
  message?: string;
  details?: ApiFieldError[];
}
