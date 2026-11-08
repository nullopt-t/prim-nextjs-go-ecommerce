import { api } from "@/shared/api/client";
import type { ApiResponse, ApiPaginatedResponse, ApiMessageResponse } from "@/shared/api/types";
import type { CursorMeta } from "@/shared/lib/pagination";
import type { Notification, NotificationCount } from "@/modules/notifications/types";

export type { Notification, NotificationItem, NotificationCount } from "@/modules/notifications/types";

// Backwards-compatible aliases
export type NotificationResponse = ApiPaginatedResponse<Notification, CursorMeta>;
export type UnreadCountResponse = ApiResponse<NotificationCount>;

export const notificationService = {
  getNotifications: (params?: { cursor?: string; pageSize?: number; unreadOnly?: boolean }) => {
    const query = new URLSearchParams();
    if (params?.cursor) query.set("cursor", params.cursor);
    if (params?.pageSize) query.set("pageSize", String(params.pageSize));
    if (params?.unreadOnly) query.set("unreadOnly", "true");
    const qs = query.toString();
    return api.get<ApiPaginatedResponse<Notification, CursorMeta>>(
      `/api/v1/user/notifications${qs ? `?${qs}` : ""}`
    );
  },
  getUnreadCount: () =>
    api.get<ApiResponse<NotificationCount>>("/api/v1/user/notifications/unread-count"),
  markAsRead: (id: string) =>
    api.patch<ApiMessageResponse>(`/api/v1/user/notifications/${id}/read`),
  markAllAsRead: () =>
    api.patch<ApiMessageResponse>("/api/v1/user/notifications/mark-all-read"),
  deleteNotification: (id: string) =>
    api.delete<ApiMessageResponse>(`/api/v1/user/notifications/${id}`),
};
