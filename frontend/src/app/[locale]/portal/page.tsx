import { PortalManager } from "@/features/portal/portalManager";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "PRIM Route Manager | Multi-Portal Architecture",
  description: "Select workspace environment: Public Storefront, Vendor Hub, or Admin Control Panel",
};

export default function PortalPage() {
  return <PortalManager />;
}
