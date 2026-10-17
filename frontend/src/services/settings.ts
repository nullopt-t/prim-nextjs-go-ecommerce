import { api } from "@/api/client";

export interface PublicSettings {
  storeName: string;
  storeEmail: string;
  storePhone: string;
  storeAddress: string;
  currency: string;
  timezone: string;
  defaultLocale: string;
  announcementEnabled: boolean;
  announcementText: string;
  flatRateShipping: number;
  freeShippingThreshold: number;
  estimatedDeliveryDays: string;
  allowCod: boolean;
  codFee: number;
  allowCards: boolean;
  taxRatePercent: number;
  pricesIncludeTax: boolean;
  maintenanceMode: boolean;
  maintenanceMessage: string;
}

export const settingsService = {
  /**
   * Fetch the non-sensitive storefront settings (identity, currency,
   * announcement, shipping rules, payment availability, maintenance state).
   */
  getPublicSettings: async (): Promise<{ data: PublicSettings }> => {
    return api.get<{ data: PublicSettings }>("/api/v1/settings/public");
  },
};
