import { api } from "@/shared/api/client";
import type { ApiResponse, ApiPaginatedResponse, ApiMessageResponse } from "@/shared/api/types";
import type { Notification, NotificationCount } from "@/modules/notifications/types";

export type { Notification, NotificationItem, NotificationCount } from "@/modules/notifications/types";

// Backwards-compatible aliases
export type NotificationResponse = ApiPaginatedResponse<Notification>;
export type UnreadCountResponse = ApiResponse<NotificationCount>;

export const notificationService = {
  getNotifications: (page = 1, pageSize = 15, unreadOnly = false) =>
    api.get<ApiPaginatedResponse<Notification>>(
      `/api/v1/user/notifications?page=${page}&pageSize=${pageSize}&unreadOnly=${unreadOnly}`
    ),
  getUnreadCount: () =>
    api.get<ApiResponse<NotificationCount>>("/api/v1/user/notifications/unread-count"),
  markAsRead: (id: string) =>
    api.patch<ApiMessageResponse>(`/api/v1/user/notifications/${id}/read`),
  markAllAsRead: () =>
    api.patch<ApiMessageResponse>("/api/v1/user/notifications/mark-all-read"),
  deleteNotification: (id: string) =>
    api.delete<ApiMessageResponse>(`/api/v1/user/notifications/${id}`),
};
