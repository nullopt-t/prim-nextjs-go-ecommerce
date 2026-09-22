export interface Product {
  id: string | number;
  title: string;
  price: number;
  description?: string;
  category?: string;
  image?: string;
  rating?: {
    rate: number;
    count: number;
  };
  discount?: number;
}

export interface UserProfile {
  id: string;
  name: string;
  email: string;
  phone?: string;
  avatar?: string;
}

export interface CartItem {
  id: string | number;
  product: Product;
  quantity: number;
}
