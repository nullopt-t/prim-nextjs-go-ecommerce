import React from "react";
import MainLayout from "@/components/layouts/mainLayout";
import { MaintenanceGate } from "@/components/ui";
import Recently from "@/features/home/components/ui/recently";

export default function ShopLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <MaintenanceGate>
      <MainLayout recently={<Recently />}>
        {children}
      </MainLayout>
    </MaintenanceGate>
  );
}
