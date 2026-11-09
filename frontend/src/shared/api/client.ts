import type { ApiFieldError } from "@/shared/api/types";
import { getErrorMessageByCode } from "@/shared/i18n/errors";

export const SELLER_NOT_AUTHORIZED_EVENT = "prim:seller-not-authorized";

export class ApiError extends Error {
  status: number;
  code?: string;
  details?: ApiFieldError[];

  constructor(
    message: string,
    status: number,
    code?: string,
    details?: ApiFieldError[],
  ) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

export interface ApiRequestOptions extends RequestInit {
  /**
   * Suppress the global sign-in redirect for background telemetry. The caller
   * still receives the 401, but an expired session does not interrupt the page.
   */
  skipAuthRedirect?: boolean;
}

export const getBaseUrl = () => {
  // One env variable for the API base URL: NEXT_PUBLIC_API_URL. It is available
  // on both the server and the browser, so client and server API calls use the
  // same value. The browser calls the API on its own origin, with credentialed
  // requests, so the backend must allow this origin via CORS.
  return process.env.NEXT_PUBLIC_API_URL || "http://localhost:8081";
};

const getActiveLocale = () => {
  if (typeof window === "undefined") return "en";
  const path = window.location.pathname;
  if (path.startsWith("/ar") || path === "/ar") return "ar";
  return "en";
};

function handleAuthRedirect() {
  if (typeof window === "undefined") return;
  const path = window.location.pathname;
  // Prevent redirect loops if already on auth pages
  if (path.includes("/auth")) return;

  const locale = getActiveLocale();
  const authUrl = locale === "ar" ? "/ar/auth" : "/auth";
  const returnParam = encodeURIComponent(window.location.pathname + window.location.search);
  window.location.href = `${authUrl}?redirect=${returnParam}`;
}

export async function request<T = unknown>(
  endpoint: string,
  options: ApiRequestOptions = {}
): Promise<T> {
  const url = endpoint.startsWith("http") ? endpoint : `${getBaseUrl()}${endpoint}`;
  
  const locale = getActiveLocale();
  const { skipAuthRedirect = false, headers, ...fetchOptions } = options;

  const response = await fetch(url, {
    ...fetchOptions,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": locale,
      ...headers,
    },
  });

  if (response.status === 401 && !skipAuthRedirect) {
    const isAuthEndpoint =
      endpoint.includes("/auth/me") ||
      endpoint.includes("/auth/challenge") ||
      endpoint.includes("/auth/logout");

    // Do not hijack browser navigation on probe/auth checks
    if (!isAuthEndpoint) {
      handleAuthRedirect();
    }
  }

  if (!response.ok) {
    const errorData = (await response.json().catch(() => ({}))) as {
      error?: string;
      message?: string;
      code?: string;
      details?: ApiFieldError[];
    };
    if (
      errorData.code === "SELLER_NOT_AUTHORIZED_FOR_ACTION" &&
      typeof window !== "undefined"
    ) {
      window.dispatchEvent(new Event(SELLER_NOT_AUTHORIZED_EVENT));
    }
    const translated = getErrorMessageByCode(errorData.code, locale);
    throw new ApiError(
      translated || errorData.error || errorData.message || `HTTP Error ${response.status}`,
      response.status,
      errorData.code,
      errorData.details,
    );
  }

  if (response.status === 204) return null as T;

  const text = await response.text();
  return text ? (JSON.parse(text) as T) : (null as T);
}

export const api = {
  get: <T = unknown>(endpoint: string, options?: ApiRequestOptions) =>
    request<T>(endpoint, { method: "GET", ...options }),
  post: <T = unknown, B = unknown>(endpoint: string, payload?: B, options?: ApiRequestOptions) =>
    request<T>(endpoint, {
      method: "POST",
      body: payload !== undefined ? JSON.stringify(payload) : undefined,
      ...options,
    }),
  patch: <T = unknown, B = unknown>(endpoint: string, payload?: B, options?: ApiRequestOptions) =>
    request<T>(endpoint, {
      method: "PATCH",
      body: payload !== undefined ? JSON.stringify(payload) : undefined,
      ...options,
    }),
  put: <T = unknown, B = unknown>(endpoint: string, payload?: B, options?: ApiRequestOptions) =>
    request<T>(endpoint, {
      method: "PUT",
      body: payload !== undefined ? JSON.stringify(payload) : undefined,
      ...options,
    }),
  delete: <T = unknown>(endpoint: string, options?: ApiRequestOptions) =>
    request<T>(endpoint, { method: "DELETE", ...options }),
  /**
   * Fetches a binary/CSV endpoint with credentials and saves it to disk via an
   * object URL. Reuses the same base URL and auth handling as `request`.
   */
  download: async (endpoint: string, filename: string, options?: RequestInit): Promise<void> => {
    const url = endpoint.startsWith("http") ? endpoint : `${getBaseUrl()}${endpoint}`;
    const locale = getActiveLocale();

    const response = await fetch(url, {
      credentials: "include",
      headers: {
        "Accept-Language": locale,
        ...options?.headers,
      },
      ...options,
    });

    if (response.status === 401) {
      handleAuthRedirect();
    }

    if (!response.ok) {
      const errorData = (await response.json().catch(() => ({}))) as {
        error?: string;
        message?: string;
        code?: string;
        details?: ApiFieldError[];
      };
      throw new ApiError(
        errorData.error || errorData.message || `HTTP Error ${response.status}`,
        response.status,
        errorData.code,
        errorData.details,
      );
    }

    const blob = await response.blob();
    const objectUrl = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = objectUrl;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    link.remove();
    URL.revokeObjectURL(objectUrl);
  },
  upload: <T = unknown>(endpoint: string, formData: FormData, options?: RequestInit) => {
    const url = endpoint.startsWith("http") ? endpoint : `${getBaseUrl()}${endpoint}`;
    const locale = getActiveLocale();
    return fetch(url, {
      method: "POST",
      credentials: "include",
      headers: {
        "Accept-Language": locale,
        ...options?.headers,
      },
      body: formData,
      ...options,
    }).then(async (res) => {
      if (!res.ok) {
        const errorData = (await res.json().catch(() => ({}))) as {
          error?: string;
          message?: string;
          code?: string;
          details?: ApiFieldError[];
        };
        throw new ApiError(
          errorData.error || errorData.message || `HTTP Error ${res.status}`,
          res.status,
          errorData.code,
          errorData.details,
        );
      }
      return res.json() as Promise<T>;
    });
  },
};

export default api;
