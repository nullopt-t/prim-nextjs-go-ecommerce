import React from "react";
import { Header, Footer } from "@/components/ui";

interface MainLayoutProps {
  children: React.ReactNode;
  recently?: React.ReactNode;
}

export function MainLayout({ children, recently }: MainLayoutProps) {
  return (
    <div className="flex flex-col min-h-screen bg-background text-foreground">
      <Header />
      <main className="w-full px-3 sm:px-6 lg:px-8 py-4 sm:py-6 flex-1">
        {children}
      </main>
      {recently}
      <Footer />
    </div>
  );
}

export default MainLayout;
