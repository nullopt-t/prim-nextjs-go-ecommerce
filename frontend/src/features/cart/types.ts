export interface CartItemData {
  id: string;
  productId?: string;
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
}
