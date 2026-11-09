import { Suspense } from "react";
import { AdminOrdersSearch } from "@/modules/admin/routes/ordersSearch";

export default function AdminOrdersSearchPage() {
  return (
    <Suspense fallback={null}>
      <AdminOrdersSearch />
    </Suspense>
  );
}
