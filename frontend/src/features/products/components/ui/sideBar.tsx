"use client";

import SideBarCategories from "@/features/products/components/ui/sideBarCategories";
import SideBarDiscount from "@/features/products/components/ui/sideBarDiscount";
import SideBarRatings from "@/features/products/components/ui/sideBarRatings";
import SideBarFilters from "@/features/products/components/ui/sideBarFilters";
import SideBarAvailability from "@/features/products/components/ui/sideBarAvailability";
import SideBarBrands from "@/features/products/components/ui/sideBarBrands";
import { PanelLeftOpen, PanelLeftClose, SlidersHorizontal, X } from "lucide-react";
import { useState, useEffect } from "react";

export default function SideBar() {
  const [isOpenDesktop, setIsOpenDesktop] = useState(true);
  const [isOpenMobile, setIsOpenMobile] = useState(false);

  // Lock body scroll when mobile filter drawer is open
  useEffect(() => {
    if (isOpenMobile) {
      document.body.style.overflow = "hidden";
    } else {
      document.body.style.overflow = "";
    }
    return () => {
      document.body.style.overflow = "";
    };
  }, [isOpenMobile]);

  const filterContent = (
    <div className="flex flex-col gap-5">
      <SideBarFilters />
      <SideBarCategories />
      <SideBarBrands />
      <SideBarRatings />
      <SideBarAvailability />
      <SideBarDiscount />
    </div>
  );

  return (
    <>
      {/* Mobile filter trigger button */}
      <div className="block md:hidden mb-4">
        <button
          type="button"
          onClick={() => setIsOpenMobile(true)}
          className="w-full flex items-center justify-center gap-2 py-3 px-4 rounded-xl bg-card border border-border text-foreground font-semibold text-sm shadow-xs hover:border-accent-brand hover:text-accent-brand transition-colors cursor-pointer"
        >
          <SlidersHorizontal className="size-4" />
          <span>Filter & Refine</span>
        </button>
      </div>

      {/* Mobile Filter Drawer / Modal */}
      {isOpenMobile && (
        <div className="fixed inset-0 z-50 md:hidden flex flex-col justify-end bg-black/60 backdrop-blur-xs animate-in fade-in duration-200">
          <div
            className="w-full max-h-[85vh] bg-background border-t border-border rounded-t-3xl p-6 flex flex-col shadow-2xl animate-in slide-in-from-bottom duration-300"
            onClick={(e) => e.stopPropagation()}
          >
            {/* Header */}
            <div className="flex items-center justify-between pb-4 border-b border-border">
              <div className="flex items-center gap-2">
                <SlidersHorizontal className="size-5 text-accent-brand" />
                <h3 className="text-lg font-bold text-foreground">Filters</h3>
              </div>
              <button
                type="button"
                onClick={() => setIsOpenMobile(false)}
                className="p-1 rounded-lg text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors cursor-pointer"
              >
                <X className="size-5" />
              </button>
            </div>

            {/* Scrollable Filter List */}
            <div className="overflow-y-auto py-4 flex-1">
              {filterContent}
            </div>

            {/* Bottom Done button */}
            <div className="pt-3 border-t border-border">
              <button
                type="button"
                onClick={() => setIsOpenMobile(false)}
                className="w-full py-3 rounded-xl bg-accent-brand text-white font-bold text-sm hover:opacity-90 transition-opacity cursor-pointer shadow-sm"
              >
                Apply Filters
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Desktop Sidebar */}
      <div className="hidden md:block p-2.5 lg:p-5 border-r border-border min-h-full">
        <button
          type="button"
          aria-label="Toggle filters sidebar"
          className="block ml-auto text-muted-foreground mb-5 hover:text-accent-brand cursor-pointer"
          onClick={() => setIsOpenDesktop((prev) => !prev)}
        >
          {isOpenDesktop ? <PanelLeftClose className="size-5" /> : <PanelLeftOpen className="size-5" />}
        </button>

        <div
          className={`overflow-hidden transition-all duration-300 ${
            isOpenDesktop ? "w-52 opacity-100" : "w-0 opacity-0 pointer-events-none"
          }`}
        >
          {filterContent}
        </div>
      </div>
    </>
  );
}
