// Centralized mock data and local storage state for Vendor Portal

export interface Vendor {
  id: string;
  name: string;
  owner?: string;
  email: string;
  rating?: number;
  commissionRate: number;
  balance?: number;
  status?: string;
  totalSales?: number;
  joinedDate?: string;
}

export const INITIAL_VENDORS: Vendor[] = [
  { id: "v-1", name: "SoundMaster Studio", owner: "Alex Turner", email: "alex@soundmaster.audio", rating: 4.9, commissionRate: 10 },
  { id: "v-2", name: "KeyCrafters Supply", owner: "Elena Rostova", email: "elena@keycrafters.dev", rating: 4.8, commissionRate: 12 },
  { id: "v-3", name: "UrbanCarry Goods", owner: "Marcus Brody", email: "marcus@urbancarry.com", rating: 4.7, commissionRate: 10 }
];

export type VendorProduct = {
  id: string;
  vendorId: string;
  vendorName: string;
  name: string;
  sku: string;
  category: string;
  price: number;
  cost: number;
  stock: number;
  status: string;
  salesCount: number;
  image: string;
};

export const INITIAL_VENDOR_PRODUCTS: VendorProduct[] = [
  {
    id: "vprod-1",
    vendorId: "v-1",
    vendorName: "SoundMaster Studio",
    name: "Wireless Noise-Canceling Headphones",
    sku: "SM-NC-HD1",
    category: "Electronics",
    price: 199.99,
    cost: 110.00,
    stock: 45,
    status: "Active",
    salesCount: 142,
    image: "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=400&q=80"
  },
  {
    id: "vprod-2",
    vendorId: "v-1",
    vendorName: "SoundMaster Studio",
    name: "Pro Studio Reference Monitor Speakers",
    sku: "SM-MON-02",
    category: "Audio",
    price: 349.00,
    cost: 200.00,
    stock: 12,
    status: "Active",
    salesCount: 64,
    image: "https://images.unsplash.com/photo-1545454675-3531b543be5d?w=400&q=80"
  },
  {
    id: "vprod-3",
    vendorId: "v-1",
    vendorName: "SoundMaster Studio",
    name: "USB Condenser Podcast Microphone",
    sku: "SM-MIC-03",
    category: "Audio",
    price: 89.99,
    cost: 40.00,
    stock: 4,
    status: "Low Stock",
    salesCount: 210,
    image: "https://images.unsplash.com/photo-1590658268037-6bf12165a8df?w=400&q=80"
  }
];

export type VendorOrder = {
  id: string;
  vendorId: string;
  orderId: string;
  customer: string;
  email: string;
  date: string;
  item: string;
  quantity: number;
  unitPrice: number;
  vendorEarnings: number;
  commissionFee: number;
  fulfillmentStatus: string;
  payoutStatus: string;
};

export const INITIAL_VENDOR_ORDERS: VendorOrder[] = [
  {
    id: "VORD-8801",
    vendorId: "v-1",
    orderId: "ORD-9821",
    customer: "Alex Morgan",
    email: "alex.morgan@example.com",
    date: "2026-08-12 14:30",
    item: "Wireless Noise-Canceling Headphones",
    quantity: 1,
    unitPrice: 199.99,
    vendorEarnings: 179.99, // After 10% platform commission
    commissionFee: 20.00,
    fulfillmentStatus: "Processing",
    payoutStatus: "Pending"
  },
  {
    id: "VORD-8802",
    vendorId: "v-1",
    orderId: "ORD-9819",
    customer: "David Kim",
    email: "dkim.tech@gmail.com",
    date: "2026-08-10 11:20",
    item: "Wireless Noise-Canceling Headphones",
    quantity: 1,
    unitPrice: 199.99,
    vendorEarnings: 179.99,
    commissionFee: 20.00,
    fulfillmentStatus: "Delivered",
    payoutStatus: "Paid"
  },
  {
    id: "VORD-8803",
    vendorId: "v-1",
    orderId: "ORD-9812",
    customer: "Sarah Jenkins",
    email: "s.jenkins@domain.com",
    date: "2026-08-08 09:15",
    item: "USB Condenser Podcast Microphone",
    quantity: 2,
    unitPrice: 89.99,
    vendorEarnings: 161.98,
    commissionFee: 18.00,
    fulfillmentStatus: "Delivered",
    payoutStatus: "Paid"
  }
];

export type VendorPayout = {
  id: string;
  vendorId: string;
  amount: string;
  method: string;
  status: string;
  requestedDate: string;
  processedDate: string;
  reference: string;
};

export const INITIAL_VENDOR_PAYOUTS: VendorPayout[] = [
  {
    id: "PO-7001",
    vendorId: "v-1",
    amount: "EGP 1,420.50",
    method: "Direct Deposit (Chase ending in 8812)",
    status: "Completed",
    requestedDate: "2026-08-01",
    processedDate: "2026-08-02",
    reference: "TRX-CHASE-99201"
  },
  {
    id: "PO-7002",
    vendorId: "v-1",
    amount: "EGP 890.00",
    method: "Direct Deposit (Chase ending in 8812)",
    status: "Completed",
    requestedDate: "2026-07-15",
    processedDate: "2026-07-16",
    reference: "TRX-CHASE-88120"
  }
];

export function getLocalVendorData<T>(key: string, fallback: T): T {
  try {
    const item = localStorage.getItem(`vendor_${key}`);
    return item ? (JSON.parse(item) as T) : fallback;
  } catch {
    return fallback;
  }
}

export function setLocalVendorData(key: string, value: unknown): void {
  try {
    localStorage.setItem(`vendor_${key}`, JSON.stringify(value));
  } catch (err) {
    console.error("Failed to save vendor state to localStorage", err);
  }
}
