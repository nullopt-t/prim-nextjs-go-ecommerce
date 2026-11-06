/** Browsing-history response models matching the Go history handlers. */

export interface ViewedProduct {
  productId: string;
  title: string;
  slug: string;
  basePrice: number;
  currency: string;
  thumbnailUrl?: string;
  viewedAt?: string;
}

export interface SyncHistoryResult {
  synced: number;
}
