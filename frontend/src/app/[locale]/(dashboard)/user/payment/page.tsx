import { Metadata } from "next";
import { CreditCard } from "lucide-react";

export const metadata: Metadata = {
  title: "Payment Methods | PRIM",
  description: "Manage your saved payment cards and billing methods",
};

export default function PaymentPage() {
  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-xl font-semibold text-foreground">Payment Methods</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Manage your payment options and billing settings.
        </p>
      </div>

      <div className="bg-card border border-border rounded-lg p-12 flex flex-col items-center justify-center text-center gap-3 shadow-sm">
        <CreditCard className="size-10 text-muted-foreground/50" />
        <p className="font-medium text-foreground">Feature Coming Soon</p>
        <p className="text-sm text-muted-foreground">
          We&apos;re working on this feature. Payment method management will be available soon.
        </p>
      </div>
    </div>
  );
}
