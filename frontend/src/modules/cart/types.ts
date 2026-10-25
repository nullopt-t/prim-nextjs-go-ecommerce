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
