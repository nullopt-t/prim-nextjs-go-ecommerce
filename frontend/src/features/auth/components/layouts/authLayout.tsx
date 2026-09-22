"use client";

import Image from "next/image";
import AuthImage from "@/assets/imgs/auth.png";

export function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 min-h-screen max-h-screen overflow-hidden bg-background text-foreground">
      <div className="flex flex-col justify-center px-6 md:px-16 overflow-y-auto max-h-screen py-10">
        {children}
      </div>
      <div className="hidden md:block relative w-full h-full bg-secondary">
        <Image
          src={AuthImage}
          alt="Authentication banner"
          fill
          priority
          sizes="(max-width: 1024px) 100vw, 50vw"
          className="object-cover object-center"
        />
      </div>
    </div>
  );
}

export default AuthLayout;
