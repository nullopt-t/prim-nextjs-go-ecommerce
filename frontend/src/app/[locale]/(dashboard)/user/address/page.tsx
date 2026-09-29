import { Metadata } from "next";
import { MapPin } from "lucide-react";

export const metadata: Metadata = {
  title: "Delivery Addresses | PRIM",
  description: "Manage shipping and delivery addresses",
};

export default function AddressPage() {
  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-xl font-semibold text-foreground">Addresses</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Manage your delivery and billing addresses.
        </p>
      </div>

      <div className="bg-card border border-border rounded-lg p-12 flex flex-col items-center justify-center text-center gap-3 shadow-sm">
        <MapPin className="size-10 text-muted-foreground/50" />
        <p className="font-medium text-foreground">Feature Coming Soon</p>
        <p className="text-sm text-muted-foreground">
          We&apos;re working on this feature. Address management will be available soon.
        </p>
      </div>
    </div>
  );
}
