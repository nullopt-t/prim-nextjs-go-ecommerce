// Centralized mock data and state store for Admin features

export const INITIAL_PRODUCTS = [
  {
    id: "prod-1",
    name: "Wireless Noise-Canceling Headphones",
    sku: "WNC-HD-001",
    category: "Electronics",
    brand: "SoundMaster",
    price: 199.99,
    compareAtPrice: 249.99,
    cost: 110.00,
    stock: 45,
    lowStockThreshold: 10,
    status: "Active",
    image: "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=400&q=80",
    variants: [
      { id: "v1-1", name: "Black / Standard", sku: "WNC-HD-001-BLK", price: 199.99, stock: 25 },
      { id: "v1-2", name: "Silver / Standard", sku: "WNC-HD-001-SLV", price: 199.99, stock: 20 }
    ],
    salesCount: 142
  },
  {
    id: "prod-2",
    name: "Ergonomic Mechanical Keyboard",
    sku: "EMK-RGB-002",
    category: "Electronics",
    brand: "KeyCrafters",
    price: 129.50,
    compareAtPrice: 149.99,
    cost: 65.00,
    stock: 8,
    lowStockThreshold: 15,
    status: "Active",
    image: "https://images.unsplash.com/photo-1587829741301-dc798b83add3?w=400&q=80",
    variants: [
      { id: "v2-1", name: "Tactile Brown Switches", sku: "EMK-RGB-002-BRN", price: 129.50, stock: 5 },
      { id: "v2-2", name: "Linear Red Switches", sku: "EMK-RGB-002-RED", price: 129.50, stock: 3 }
    ],
    salesCount: 89
  },
  {
    id: "prod-3",
    name: "Minimalist Leather Backpack",
    sku: "MLB-BRN-003",
    category: "Fashion",
    brand: "UrbanCarry",
    price: 89.00,
    compareAtPrice: 110.00,
    cost: 38.00,
    stock: 62,
    lowStockThreshold: 10,
    status: "Active",
    image: "https://images.unsplash.com/photo-1553062407-98eeb64c6a62?w=400&q=80",
    variants: [
      { id: "v3-1", name: "Tan Brown", sku: "MLB-BRN-003-TAN", price: 89.00, stock: 32 },
      { id: "v3-2", name: "Matte Black", sku: "MLB-BRN-003-BLK", price: 89.00, stock: 30 }
    ],
    salesCount: 210
  },
  {
    id: "prod-4",
    name: "Smart Fitness Watch V2",
    sku: "SFW-BLK-004",
    category: "Electronics",
    brand: "FitTech",
    price: 179.00,
    compareAtPrice: 210.00,
    cost: 85.00,
    stock: 3,
    lowStockThreshold: 10,
    status: "Low Stock",
    image: "https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=400&q=80",
    variants: [
      { id: "v4-1", name: "Black / Silicone", sku: "SFW-BLK-004-SIL", price: 179.00, stock: 3 }
    ],
    salesCount: 312
  },
  {
    id: "prod-5",
    name: "Stainless Steel Insulated Bottle",
    sku: "SIB-SLV-005",
    category: "Home & Kitchen",
    brand: "HydroGear",
    price: 29.99,
    compareAtPrice: 35.00,
    cost: 8.50,
    stock: 120,
    lowStockThreshold: 20,
    status: "Active",
    image: "https://images.unsplash.com/photo-1602143407151-7111542de6e8?w=400&q=80",
    variants: [
      { id: "v5-1", name: "750ml / Matte White", sku: "SIB-SLV-005-WHT", price: 29.99, stock: 60 },
      { id: "v5-2", name: "750ml / Deep Blue", sku: "SIB-SLV-005-BLU", price: 29.99, stock: 60 }
    ],
    salesCount: 540
  }
];

export const INITIAL_CATEGORIES = [
  { id: "cat-1", name: "Electronics", slug: "electronics", parent: "None", productsCount: 24, status: "Active", icon: "Cpu" },
  { id: "cat-2", name: "Fashion & Apparel", slug: "fashion", parent: "None", productsCount: 48, status: "Active", icon: "ShoppingBag" },
  { id: "cat-3", name: "Home & Kitchen", slug: "home-kitchen", parent: "None", productsCount: 32, status: "Active", icon: "Home" },
  { id: "cat-4", name: "Audio Devices", slug: "audio", parent: "Electronics", productsCount: 12, status: "Active", icon: "Headphones" },
  { id: "cat-5", name: "Wearables", slug: "wearables", parent: "Electronics", productsCount: 8, status: "Active", icon: "Watch" }
];

export const INITIAL_BRANDS = [
  { id: "brand-1", name: "SoundMaster", website: "https://soundmaster.audio", productsCount: 8, status: "Active", logo: "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=100&q=80" },
  { id: "brand-2", name: "KeyCrafters", website: "https://keycrafters.dev", productsCount: 5, status: "Active", logo: "https://images.unsplash.com/photo-1587829741301-dc798b83add3?w=100&q=80" },
  { id: "brand-3", name: "UrbanCarry", website: "https://urbancarry.com", productsCount: 14, status: "Active", logo: "https://images.unsplash.com/photo-1553062407-98eeb64c6a62?w=100&q=80" },
  { id: "brand-4", name: "FitTech", website: "https://fittechwear.io", productsCount: 6, status: "Active", logo: "https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=100&q=80" },
  { id: "brand-5", name: "HydroGear", website: "https://hydrogear.store", productsCount: 9, status: "Active", logo: "https://images.unsplash.com/photo-1602143407151-7111542de6e8?w=100&q=80" }
];

export const INITIAL_ATTRIBUTES = [
  { id: "attr-1", name: "Color", code: "color", values: ["Black", "Silver", "Tan Brown", "Deep Blue", "Matte White"] },
  { id: "attr-2", name: "Size", code: "size", values: ["Small", "Medium", "Large", "750ml", "1000ml"] },
  { id: "attr-3", name: "Switch Type", code: "switch_type", values: ["Linear Red", "Tactile Brown", "Clicky Blue"] },
  { id: "attr-4", name: "Material", code: "material", values: ["Genuine Leather", "Aluminum", "Stainless Steel", "Silicone"] }
];

export const INITIAL_MEDIA = [
  { id: "med-1", name: "headphones-black.jpg", size: "1.2 MB", type: "Image", dimensions: "1200x1200", url: "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=800&q=80", uploadedAt: "2026-08-10" },
  { id: "med-2", name: "keyboard-rgb.jpg", size: "2.4 MB", type: "Image", dimensions: "1600x1200", url: "https://images.unsplash.com/photo-1587829741301-dc798b83add3?w=800&q=80", uploadedAt: "2026-08-09" },
  { id: "med-3", name: "leather-backpack.jpg", size: "1.8 MB", type: "Image", dimensions: "1200x1600", url: "https://images.unsplash.com/photo-1553062407-98eeb64c6a62?w=800&q=80", uploadedAt: "2026-08-08" },
  { id: "med-4", name: "smartwatch-front.jpg", size: "950 KB", type: "Image", dimensions: "1000x1000", url: "https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=800&q=80", uploadedAt: "2026-08-07" },
  { id: "med-5", name: "hydro-bottle.jpg", size: "1.1 MB", type: "Image", dimensions: "1200x1200", url: "https://images.unsplash.com/photo-1602143407151-7111542de6e8?w=800&q=80", uploadedAt: "2026-08-06" }
];

export const INITIAL_ORDERS = [
  {
    id: "ORD-9821",
    customer: "Alex Morgan",
    email: "alex.morgan@example.com",
    date: "2026-08-12 14:30",
    total: "$288.99",
    amountNum: 288.99,
    status: "Processing",
    paymentStatus: "Paid",
    fulfillmentStatus: "Unfulfilled",
    paymentMethod: "Credit Card (Visa ending in 4242)",
    shippingAddress: "742 Evergreen Terrace, Springfield, OR 97477",
    items: [
      { id: "item-1", name: "Wireless Noise-Canceling Headphones", price: "$199.99", quantity: 1 },
      { id: "item-2", name: "Minimalist Leather Backpack", price: "$89.00", quantity: 1 }
    ],
    trackingNumber: "TRK-9821-EXP",
    history: [
      { date: "2026-08-12 14:30", note: "Order placed by customer." },
      { date: "2026-08-12 14:31", note: "Payment of $288.99 confirmed via Stripe." }
    ]
  },
  {
    id: "ORD-9820",
    customer: "Sarah Jenkins",
    email: "s.jenkins@domain.com",
    date: "2026-08-11 18:15",
    total: "$129.50",
    amountNum: 129.50,
    status: "Shipped",
    paymentStatus: "Paid",
    fulfillmentStatus: "Fulfilled",
    paymentMethod: "PayPal (s.jenkins@domain.com)",
    shippingAddress: "100 Universal City Plaza, Universal City, CA 91608",
    items: [
      { id: "item-3", name: "Ergonomic Mechanical Keyboard", price: "$129.50", quantity: 1 }
    ],
    trackingNumber: "FEDEX-8891023",
    history: [
      { date: "2026-08-11 18:15", note: "Order placed." },
      { date: "2026-08-11 18:16", note: "Payment captured via PayPal." },
      { date: "2026-08-12 09:00", note: "Package handed to FedEx. Tracking: FEDEX-8891023." }
    ]
  },
  {
    id: "ORD-9819",
    customer: "David Kim",
    email: "dkim.tech@gmail.com",
    date: "2026-08-10 11:20",
    total: "$387.99",
    amountNum: 387.99,
    status: "Delivered",
    paymentStatus: "Paid",
    fulfillmentStatus: "Fulfilled",
    paymentMethod: "Apple Pay",
    shippingAddress: "555 California St, San Francisco, CA 94104",
    items: [
      { id: "item-4", name: "Smart Fitness Watch V2", price: "$179.00", quantity: 1 },
      { id: "item-5", name: "Wireless Noise-Canceling Headphones", price: "$199.99", quantity: 1 },
      { id: "item-6", name: "Stainless Steel Insulated Bottle", price: "$29.99", quantity: 1 }
    ],
    trackingNumber: "UPS-1Z9999999999",
    history: [
      { date: "2026-08-10 11:20", note: "Order created." },
      { date: "2026-08-10 14:00", note: "Shipped via UPS." },
      { date: "2026-08-11 16:30", note: "Delivered to front porch." }
    ]
  },
  {
    id: "ORD-9818",
    customer: "Elena Rostova",
    email: "elena.r@design.co",
    date: "2026-08-09 09:45",
    total: "$59.98",
    amountNum: 59.98,
    status: "Cancelled",
    paymentStatus: "Refunded",
    fulfillmentStatus: "Unfulfilled",
    paymentMethod: "Credit Card (Mastercard ending in 1102)",
    shippingAddress: "200 Park Ave, New York, NY 10166",
    items: [
      { id: "item-7", name: "Stainless Steel Insulated Bottle", price: "$29.99", quantity: 2 }
    ],
    trackingNumber: "N/A",
    history: [
      { date: "2026-08-09 09:45", note: "Order placed." },
      { date: "2026-08-09 10:15", note: "Customer requested cancellation." },
      { date: "2026-08-09 10:20", note: "Full refund of $59.98 issued." }
    ]
  }
];

export const INITIAL_CUSTOMERS = [
  { id: "cust-1", name: "Alex Morgan", email: "alex.morgan@example.com", phone: "+1 (555) 234-5678", ordersCount: 4, totalSpent: "$682.50", segment: "VIP", status: "Active", joinedDate: "2025-11-12" },
  { id: "cust-2", name: "Sarah Jenkins", email: "s.jenkins@domain.com", phone: "+1 (555) 876-5432", ordersCount: 2, totalSpent: "$240.00", segment: "Regular", status: "Active", joinedDate: "2026-01-05" },
  { id: "cust-3", name: "David Kim", email: "dkim.tech@gmail.com", phone: "+1 (555) 432-1098", ordersCount: 7, totalSpent: "$1,450.00", segment: "VIP", status: "Active", joinedDate: "2025-08-20" },
  { id: "cust-4", name: "Elena Rostova", email: "elena.r@design.co", phone: "+1 (555) 654-9870", ordersCount: 1, totalSpent: "$59.98", segment: "New", status: "Active", joinedDate: "2026-08-01" },
  { id: "cust-5", name: "Marcus Brody", email: "mbrody@archeology.edu", phone: "+1 (555) 998-1122", ordersCount: 0, totalSpent: "$0.00", segment: "New", status: "Inactive", joinedDate: "2026-08-10" }
];

export const INITIAL_PAYMENTS = [
  { id: "PAY-501", orderId: "ORD-9821", customer: "Alex Morgan", gateway: "Stripe (Credit Card)", amount: "$288.99", fee: "$8.68", status: "Succeeded", date: "2026-08-12 14:31" },
  { id: "PAY-502", orderId: "ORD-9820", customer: "Sarah Jenkins", gateway: "PayPal", amount: "$129.50", fee: "$3.88", status: "Succeeded", date: "2026-08-11 18:16" },
  { id: "PAY-503", orderId: "ORD-9819", customer: "David Kim", gateway: "Apple Pay", amount: "$387.99", fee: "$11.64", status: "Succeeded", date: "2026-08-10 11:20" },
  { id: "PAY-504", orderId: "ORD-9818", customer: "Elena Rostova", gateway: "Stripe (Credit Card)", amount: "$59.98", fee: "$0.00", status: "Refunded", date: "2026-08-09 10:20" }
];

export const INITIAL_PROMOTIONS = [
  { id: "promo-1", code: "SUMMER20", type: "Percentage", value: "20% OFF", usageLimit: 500, currentUsage: 214, status: "Active", startDate: "2026-06-01", endDate: "2026-08-31" },
  { id: "promo-2", code: "WELCOME10", type: "Fixed Amount", value: "$10.00 OFF", usageLimit: 1000, currentUsage: 642, status: "Active", startDate: "2026-01-01", endDate: "2026-12-31" },
  { id: "promo-3", code: "FREESHIP", type: "Free Shipping", value: "Free Standard Shipping", usageLimit: 200, currentUsage: 189, status: "Active", startDate: "2026-08-01", endDate: "2026-08-20" },
  { id: "promo-4", code: "FLASH50", type: "Percentage", value: "50% OFF", usageLimit: 50, currentUsage: 50, status: "Expired", startDate: "2026-07-15", endDate: "2026-07-16" }
];

export const INITIAL_INVENTORY_LOGS = [
  { id: "inv-log-1", product: "Smart Fitness Watch V2", sku: "SFW-BLK-004-SIL", change: "-5", type: "Order Fulfillment", newStock: 3, date: "2026-08-12 14:30", user: "System" },
  { id: "inv-log-2", product: "Wireless Headphones", sku: "WNC-HD-001-BLK", change: "+50", type: "Restock", newStock: 25, date: "2026-08-10 09:00", user: "Warehouse Admin" },
  { id: "inv-log-3", product: "Ergonomic Keyboard", sku: "EMK-RGB-002-BRN", change: "-2", type: "Damaged Adjustment", newStock: 5, date: "2026-08-08 16:20", user: "Store Manager" }
];

export const INITIAL_SETTINGS = {
  storeName: "PRIM Global Store",
  storeEmail: "support@prim.com",
  currency: "USD ($)",
  timezone: "UTC-7 (Pacific Time)",
  taxRatePercent: 8.5,
  freeShippingThreshold: 150,
  flatRateShipping: 9.99,
  orderNotifications: true,
  lowStockAlerts: true,
  enableGuestCheckout: true,
  apiKey: "pk_live_prim_99812401824710129",
  stripeWebhookSecret: "whsec_prim_09182391283"
};

// Helper hook / helper loader to manage local state
export function getLocalAdminData(key, fallback) {
  try {
    const item = localStorage.getItem(`admin_${key}`);
    return item ? JSON.parse(item) : fallback;
  } catch {
    return fallback;
  }
}

export function setLocalAdminData(key, value) {
  try {
    localStorage.setItem(`admin_${key}`, JSON.stringify(value));
  } catch (e) {
    console.error("Failed to save admin state to localStorage", e);
  }
}
