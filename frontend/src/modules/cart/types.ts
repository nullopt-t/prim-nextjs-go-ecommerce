/**
 * Cart domain and response models matching backend CartResponse.
 */

export interface CartSummary {
  subtotal: number;
  discount: number;
  shipping: number;
  tax: number;
  total: number;
}

export interface CartItem {
  id: string;
  productId: string;
  productSlug?: string;
  variantId: string;
  color?: string;
  attributes?: Record<string, unknown>;
  title: string;
  thumbnailUrl: string;
  quantity: number;
  unitPrice: number;        // in cents
  subtotal: number;         // in cents
  inStock: boolean;
  availableStock: number;
}

export interface Cart {
  id: string;
  currency: string;
  items: CartItem[];
  summary: CartSummary;
  itemCount: number;
  updatedAt: string;
}

export interface AddItemPayload {
  variantId?: string;
  id?: string;
  slug?: string;
  quantity?: number;
}

export interface UpdateItemQuantityPayload {
  quantity: number;
}

export interface CouponApplication {
  id?: string;
  code: string;
  discountType?: "percentage" | "fixed_amount" | string;
  discountValue?: number;
  minCartAmount?: number;
  discountAmount?: number;
  subtotal?: number;
  shipping?: number;
  total?: number;
  freeShipping?: boolean;
  message?: string;
}

/** Legacy UI format for OrderBox component */
export interface CartItemData {
  id: string;
  productId?: string;
  productSlug?: string;
  variantId?: string;
  productName: {
    en?: string;
    ar?: string;
  } | string;
  productBrand?: string;
  productPrice: string | number;
  quantity?: number;
  img?: string;
  color?: string;
  inStock?: boolean;
  availableStock?: number;
}
