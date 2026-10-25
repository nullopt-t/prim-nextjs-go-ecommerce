/**
 * Notification domain models matching backend user/notification_handler.go.
 */

export interface Notification {
  id: string;
  title: string;
  message: string;
  category: string;
  actionUrl?: string;
  metadata?: Record<string, unknown>;
  isRead: boolean;
  readAt?: string;
  createdAt: string;
}

export type NotificationItem = Notification;

export interface NotificationCount {
  count: number;
}
