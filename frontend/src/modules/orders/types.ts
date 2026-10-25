/**
 * Order domain and response models matching backend OrderResponse.
 */

export interface OrderItem {
  id: string;
  variantId: string;
  quantity: number;
  priceAtPurchase: number;    // in cents
  productSnapshot?: string;
  // Resolved storefront properties if hydrated
  title?: string;
  thumbnailUrl?: string;
}

export interface OrderAddress {
  street?: string;
  city?: string;
  state?: string;
  postalCode?: string;
  country?: string;
}

export type OrderStatus =
  | "pending"
  | "processing"
  | "shipped"
  | "delivered"
  | "cancelled";

export interface Order {
  id: string;
  customerId?: string;
  customerEmail: string;
  shippingAddress?: OrderAddress;
  status: OrderStatus;
  couponId?: string | null;
  couponCode?: string | null;
  discountAmount: number;   // in cents
  shippingAmount?: number;  // in cents
  taxAmount?: number;       // in cents
  codFee?: number;          // in cents
  totalAmount: number;      // in cents
  currency: string;
  paymentMethod?: string;
  paymentStatus?: string;
  trackingNumber?: string;
  items: OrderItem[];
  createdAt: string;
  updatedAt: string;
}

export interface CheckoutPayload {
  shippingAddressId: string;
  couponId?: string;
  paymentMethod: string;
}

export interface CheckoutResult {
  orderId: string;
  status: string;
  paymentStatus?: string;
}

export interface UpdateOrderPayload {
  status?: OrderStatus | string;
  notes?: string;
  trackingNumber?: string;
}
