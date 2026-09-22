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
      <main className="p-2.5 pt-0 md:p-5 md:pt-0 lg:pt-0 lg:p-10 flex-1">
        {children}
      </main>
      {recently}
      <Footer />
    </div>
  );
}

export default MainLayout;
