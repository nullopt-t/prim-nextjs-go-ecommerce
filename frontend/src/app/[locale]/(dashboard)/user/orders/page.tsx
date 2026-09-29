import { Metadata } from "next";
import { Package } from "lucide-react";
import Link from "next/link";

export const metadata: Metadata = {
  title: "My Orders | PRIM",
  description: "View and track your previous orders",
};

export default function OrdersPage() {
  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-xl font-semibold text-foreground">My Orders</h1>
        <p className="text-sm text-muted-foreground mt-1">
          View and track all your recent orders.
        </p>
      </div>

      <div className="bg-card border border-border rounded-lg p-12 flex flex-col items-center justify-center text-center gap-3 shadow-sm">
        <Package className="size-10 text-muted-foreground/50" />
        <p className="font-medium text-foreground">No orders yet</p>
        <p className="text-sm text-muted-foreground">
          Your orders will appear here once you&apos;ve made a purchase.
        </p>
        <Link
          href="/products"
          className="mt-2 h-9 px-4 inline-flex items-center rounded-md bg-accent-brand text-white text-sm font-medium hover:bg-accent-brand/90 transition shadow-sm"
        >
          Start Shopping
        </Link>
      </div>
    </div>
  );
}
