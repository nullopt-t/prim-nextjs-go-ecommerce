import { Metadata } from "next";

export const metadata: Metadata = {
  title: "My Orders | PRIM",
  description: "View and track your previous orders",
};

export default function OrdersPage() {
  return (
    <div>
      <h1 className="font-medium text-title-sm md:text-title-md mb-2 text-foreground">
        My Orders
      </h1>
      <p className="text-muted-foreground text-txt-sm mb-6">
        View and track all your recent orders
      </p>
      <div className="border border-border rounded-lg p-6 text-center text-muted-foreground bg-card">
        No orders found.
      </div>
    </div>
  );
}
