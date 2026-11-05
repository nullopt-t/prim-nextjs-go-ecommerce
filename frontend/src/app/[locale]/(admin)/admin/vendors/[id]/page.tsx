"use client";

import { AdminVendorDetail } from "@/modules/admin";
import { useParams } from "next/navigation";

export default function AdminVendorDetailPage() {
  const { id } = useParams();
  return <AdminVendorDetail vendorId={String(id)} />;
}
