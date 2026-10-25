const getBaseUrl = () => {
  // The browser calls the API on its own origin, with credentialed requests, so
  // the backend must allow this origin via CORS. The server prefers the
  // internal API_URL when one is configured.
  if (typeof window !== "undefined") {
    return process.env.NEXT_PUBLIC_API_URL || "http://localhost:8081";
  }
  return process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || "http://localhost:8081";
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
  options: RequestInit = {}
): Promise<T> {
  const url = endpoint.startsWith("http") ? endpoint : `${getBaseUrl()}${endpoint}`;
  
  const locale = getActiveLocale();

  const response = await fetch(url, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": locale,
      ...options.headers,
    },
    ...options,
  });

  if (response.status === 401) {
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
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || errorData.message || `HTTP Error ${response.status}`);
  }

  if (response.status === 204) return null as T;

  const text = await response.text();
  return text ? (JSON.parse(text) as T) : (null as T);
}

export const api = {
  get: <T = unknown>(endpoint: string, options?: RequestInit) =>
    request<T>(endpoint, { method: "GET", ...options }),
  post: <T = unknown, B = unknown>(endpoint: string, payload?: B, options?: RequestInit) =>
    request<T>(endpoint, {
      method: "POST",
      body: payload !== undefined ? JSON.stringify(payload) : undefined,
      ...options,
    }),
  patch: <T = unknown, B = unknown>(endpoint: string, payload?: B, options?: RequestInit) =>
    request<T>(endpoint, {
      method: "PATCH",
      body: payload !== undefined ? JSON.stringify(payload) : undefined,
      ...options,
    }),
  put: <T = unknown, B = unknown>(endpoint: string, payload?: B, options?: RequestInit) =>
    request<T>(endpoint, {
      method: "PUT",
      body: payload !== undefined ? JSON.stringify(payload) : undefined,
      ...options,
    }),
  delete: <T = unknown>(endpoint: string, options?: RequestInit) =>
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
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || errorData.message || `HTTP Error ${response.status}`);
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
        const errorData = await res.json().catch(() => ({}));
        throw new Error(errorData.error || errorData.message || `HTTP Error ${res.status}`);
      }
      return res.json() as Promise<T>;
    });
  },
};

export default api;
