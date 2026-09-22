import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Payment Methods | PRIM",
  description: "Manage your saved payment cards and billing methods",
};

export default function PaymentPage() {
  return (
    <div>
      <h1 className="font-medium text-title-sm md:text-title-md mb-2 text-foreground">
        Payment Methods
      </h1>
      <p className="text-muted-foreground text-txt-sm mb-6">
        Manage your payment options and billing settings
      </p>
      <div className="border border-border rounded-lg p-6 text-center text-muted-foreground bg-card">
        No payment methods added yet.
      </div>
    </div>
  );
}
