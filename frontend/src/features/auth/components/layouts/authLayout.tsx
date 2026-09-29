"use client";

import Image from "next/image";
import AuthImage from "@/assets/imgs/auth.png";
import { Sparkles } from "lucide-react";

export function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 min-h-screen bg-background text-foreground selection:bg-accent-brand/20">
      {/* Form side */}
      <div className="flex flex-col justify-center items-center px-4 sm:px-8 md:px-12 lg:px-16 py-8 sm:py-12 min-h-screen overflow-y-auto">
        <div className="w-full max-w-md">
          {/* Brand header badge */}
          <div className="flex items-center gap-2 mb-8">
            <div className="size-9 rounded-md bg-accent-brand/10 border border-accent-brand/25 flex items-center justify-center text-accent-brand">
              <Sparkles className="size-5" />
            </div>
            <span className="font-black text-xl tracking-tight bg-gradient-to-r from-foreground to-foreground/70 bg-clip-text text-transparent">
              PRIM
            </span>
          </div>

          <div className="bg-card/40 backdrop-blur-xs rounded-md border border-border/60 p-6 sm:p-8 shadow-xs">
            {children}
          </div>
        </div>
      </div>

      {/* Decorative / Brand Image Side */}
      <div className="hidden lg:relative lg:block w-full h-full min-h-screen bg-muted/30 overflow-hidden">
        <Image
          src={AuthImage}
          alt="Authentication banner"
          fill
          priority
          sizes="50vw"
          className="object-cover object-center"
        />
        <div className="absolute inset-0 bg-gradient-to-t from-background/80 via-transparent to-transparent pointer-events-none" />
      </div>
    </div>
  );
}

export default AuthLayout;
