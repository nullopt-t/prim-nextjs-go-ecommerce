import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Delivery Addresses | PRIM",
  description: "Manage shipping and delivery addresses",
};

export default function AddressPage() {
  return (
    <div>
      <h1 className="font-medium text-title-sm md:text-title-md mb-2 text-foreground">
        Addresses
      </h1>
      <p className="text-muted-foreground text-txt-sm mb-6">
        Manage your delivery and billing addresses
      </p>
      <div className="border border-border rounded-lg p-6 text-center text-muted-foreground bg-card">
        No addresses saved yet.
      </div>
    </div>
  );
}
