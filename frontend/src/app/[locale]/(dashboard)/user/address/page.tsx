import { Metadata } from "next";
import { AddressesContent } from "@/features/user";

export const metadata: Metadata = {
  title: "Delivery Addresses | PRIM",
  description: "Manage shipping and delivery addresses",
};

export default function AddressPage() {
  return <AddressesContent />;
}
