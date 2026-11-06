"use client";

import { AdminSellerDetail } from "@/modules/admin";
import { useParams } from "next/navigation";

export default function AdminSellerDetailPage() {
  const { id } = useParams();
  return <AdminSellerDetail sellerId={String(id)} />;
}
