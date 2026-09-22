"use client";

import SideBarCategories from "@/features/products/components/ui/sideBarCategories";
import SideBarDiscount from "@/features/products/components/ui/sideBarDiscount";
import SideBarRatings from "@/features/products/components/ui/sideBarRatings";
import SideBarFilters from "@/features/products/components/ui/sideBarFilters";
import SideBarAvailability from "@/features/products/components/ui/sideBarAvailability";
import SideBarBrands from "@/features/products/components/ui/sideBarBrands";
import { PanelLeftOpen, PanelLeftClose } from "lucide-react";
import { useState } from "react";

export default function SideBar() {
  const [isOpen, setIsOpen] = useState(true);

  return (
    <div className="p-2.5 lg:p-5 border-r border-border min-h-full">
      <button
        type="button"
        aria-label="Toggle filters sidebar"
        className="block ml-auto text-muted-foreground mb-5 hover:text-accent-brand"
        onClick={() => setIsOpen((prev) => !prev)}
      >
        {isOpen ? <PanelLeftClose className="size-5" /> : <PanelLeftOpen className="size-5" />}
      </button>

      <div
        className={`flex flex-col gap-5 overflow-hidden transition-all duration-300 ${
          isOpen ? "w-52 opacity-100" : "w-0 opacity-0 pointer-events-none"
        }`}
      >
        <SideBarFilters />
        <SideBarCategories />
        <SideBarBrands />
        <SideBarRatings />
        <SideBarAvailability />
        <SideBarDiscount />
      </div>
    </div>
  );
}
