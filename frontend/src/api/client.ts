const getBaseUrl = () => {
  // In the browser, use relative paths to leverage Next.js rewrite proxy (same-origin cookies)
  if (typeof window !== "undefined") {
    return "";
  }
  return process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || "http://localhost:8081";
};

const getSessionId = () => {
  if (typeof window === "undefined") return "";
  let sid = localStorage.getItem("prim_guest_session_id");
  if (!sid) {
    sid = "guest-" + Math.random().toString(36).substring(2, 15) + Date.now().toString(36);
    localStorage.setItem("prim_guest_session_id", sid);
  }
  return sid;
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

export async function request<T = any>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const url = endpoint.startsWith("http") ? endpoint : `${getBaseUrl()}${endpoint}`;
  
  const sessionId = getSessionId();
  const locale = getActiveLocale();

  const response = await fetch(url, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": locale,
      ...(sessionId ? { "X-Session-ID": sessionId } : {}),
      ...options.headers,
    },
    ...options,
  });

  if (response.status === 401) {
    const isAuthEndpoint =
      endpoint.includes("/auth/challenge") ||
      endpoint.includes("/auth/logout");

    // If session is expired or invalid on protected resource, redirect to login
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
  get: <T = any>(endpoint: string, options?: RequestInit) =>
    request<T>(endpoint, { method: "GET", ...options }),
  post: <T = any, B = any>(endpoint: string, payload?: B, options?: RequestInit) =>
    request<T>(endpoint, {
      method: "POST",
      body: payload !== undefined ? JSON.stringify(payload) : undefined,
      ...options,
    }),
  patch: <T = any, B = any>(endpoint: string, payload?: B, options?: RequestInit) =>
    request<T>(endpoint, {
      method: "PATCH",
      body: payload !== undefined ? JSON.stringify(payload) : undefined,
      ...options,
    }),
  put: <T = any, B = any>(endpoint: string, payload?: B, options?: RequestInit) =>
    request<T>(endpoint, {
      method: "PUT",
      body: payload !== undefined ? JSON.stringify(payload) : undefined,
      ...options,
    }),
  delete: <T = any>(endpoint: string, options?: RequestInit) =>
    request<T>(endpoint, { method: "DELETE", ...options }),
};

export default api;
