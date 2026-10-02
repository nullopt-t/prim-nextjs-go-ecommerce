import { api } from "@/api/client";

export interface NotificationItem {
  id: string;
  title: string;
  message: string;
  category: string;
  actionUrl?: string;
  metadata?: any;
  isRead: boolean;
  readAt?: string;
  createdAt: string;
}

export interface NotificationResponse {
  data: NotificationItem[];
  meta: {
    page: number;
    pageSize: number;
    totalItems: number;
    totalPages: number;
    hasPrevious: boolean;
    hasNext: boolean;
  };
}

export interface UnreadCountResponse {
  data: {
    count: number;
  };
}

export const notificationService = {
  getNotifications: (page = 1, pageSize = 15, unreadOnly = false) =>
    api.get<NotificationResponse>(
      `/api/v1/user/notifications?page=${page}&pageSize=${pageSize}&unreadOnly=${unreadOnly}`
    ),
  getUnreadCount: () =>
    api.get<UnreadCountResponse>("/api/v1/user/notifications/unread-count"),
  markAsRead: (id: string) =>
    api.patch<{ message: string }>(`/api/v1/user/notifications/${id}/read`),
  markAllAsRead: () =>
    api.patch<{ message: string }>("/api/v1/user/notifications/mark-all-read"),
  deleteNotification: (id: string) =>
    api.delete<{ message: string }>(`/api/v1/user/notifications/${id}`),
};
