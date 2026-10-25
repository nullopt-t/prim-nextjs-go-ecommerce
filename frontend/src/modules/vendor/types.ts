/**
 * Vendor domain models.
 */

export interface Vendor {
  id: string;
  name: string;
  owner?: string;
  email: string;
  phone?: string;
  storeName?: string;
  rating?: number;
  commissionRate: number;
  balance?: number;
  status?: string;
  totalSales?: number;
  joinedDate?: string;
  createdAt?: string;
}

export interface VendorPayout {
  id: string;
  vendorId?: string;
  amount: number;
  currency?: string;
  status: "pending" | "approved" | "rejected" | "paid" | string;
  method: string;
  requestedAt?: string;
  requestedDate?: string;
  processedAt?: string;
  processedDate?: string;
  reference?: string;
}

export interface VendorStats {
  totalRevenue?: number;
  totalOrders?: number;
  totalProducts?: number;
  pendingBalance?: number;
  availableBalance?: number;
  [key: string]: number | undefined;
}

export interface RequestPayoutPayload {
  vendorId?: string;
  amount: number;
  method: string;
  notes?: string;
}
