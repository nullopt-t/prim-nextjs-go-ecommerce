import React from "react";
import MainLayout from "@/components/layouts/mainLayout";
import Recently from "@/features/home/components/ui/recently";

export default function ShopLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <MainLayout recently={<Recently />}>
      {children}
    </MainLayout>
  );
}
