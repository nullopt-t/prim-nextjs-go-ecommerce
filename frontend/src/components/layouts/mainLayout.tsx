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
      <main className="max-w-screen-2xl mx-auto w-full px-4 sm:px-6 lg:px-10 py-5 sm:py-8 flex-1">
        {children}
      </main>
      {recently}
      <Footer />
    </div>
  );
}

export default MainLayout;
