"use client";

import React, { createContext, useContext, useState, useEffect } from "react";
import { vendorService } from "@/modules/vendor/api/vendorService";
import { INITIAL_VENDORS } from "../data/mockVendorData";

export interface Vendor {
  id: string;
  name: string;
  owner?: string;
  email: string;
  rating?: number;
  commissionRate: number;
  balance?: number;
  status?: string;
  totalSales?: number;
  joinedDate?: string;
}

interface VendorContextType {
  vendor: Vendor;
  selectedVendor: Vendor;
  setSelectedVendor: React.Dispatch<React.SetStateAction<Vendor>>;
  vendors: Vendor[];
  setVendors: React.Dispatch<React.SetStateAction<Vendor[]>>;
  refreshVendors: () => Promise<void>;
  loading: boolean;
}

const defaultVendor: Vendor = INITIAL_VENDORS[0];

const VendorContext = createContext<VendorContextType>({
  vendor: defaultVendor,
  selectedVendor: defaultVendor,
  setSelectedVendor: () => {},
  vendors: INITIAL_VENDORS,
  setVendors: () => {},
  refreshVendors: async () => {},
  loading: false,
});

export function VendorProvider({ children }: { children: React.ReactNode }) {
  const [vendors, setVendors] = useState<Vendor[]>(INITIAL_VENDORS);
  const [selectedVendor, setSelectedVendor] = useState<Vendor>(INITIAL_VENDORS[0]);
  const [loading, setLoading] = useState(true);

  const loadVendors = async () => {
    try {
      const res = await vendorService.getVendors();
      const data = Array.isArray(res) ? res : (res as { data?: unknown })?.data;
      const list: Vendor[] = Array.isArray(data) ? data : [];
      if (list.length > 0) {
        setVendors(list);
        setSelectedVendor((prev) => list.find((v: Vendor) => v.id === prev.id) || list[0]);
      }
    } catch {
      // Graceful fallback
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadVendors();
  }, []);

  return (
    <VendorContext.Provider
      value={{
        vendor: selectedVendor,
        selectedVendor,
        setSelectedVendor,
        vendors,
        setVendors,
        refreshVendors: loadVendors,
        loading,
      }}
    >
      {children}
    </VendorContext.Provider>
  );
}

export function useVendorContext() {
  return useContext(VendorContext);
}
