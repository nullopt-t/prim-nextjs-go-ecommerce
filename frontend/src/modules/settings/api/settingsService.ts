import { api } from "@/shared/api/client";

export interface PublicSettings {
  storeName: string;
  storeEmail: string;
  storePhone: string;
  storeAddress: string;
  currency: string;
  defaultLocale: string;
  announcementEnabled: boolean;
  announcementText: string;
  flatRateShipping: number;
  giftWrapFee?: number;
  freeShippingThreshold: number;
  /** Optional until the backend exposes a dedicated express rate. */
  expressShippingRate?: number;
  estimatedDeliveryDays: string;
  allowCod: boolean;
  codFee: number;
  allowCards: boolean;
  taxRatePercent: number;
  pricesIncludeTax: boolean;
  maintenanceMode: boolean;
  maintenanceMessage: string;
}

export interface BackendCurrency {
  code: string;
  name: string;
  symbol: string;
  flag: string;
  decimals: number;
  rateToEGP: number;
  rateToBase: number;
  isBase: boolean;
  enabled: boolean;
}

export const settingsService = {
  /**
   * Fetch the non-sensitive storefront settings (identity, currency,
   * announcement, shipping rules, payment availability, maintenance state).
   */
  getPublicSettings: async (): Promise<{ data: PublicSettings }> => {
    return api.get<{ data: PublicSettings }>("/api/v1/settings/public", {
      cache: "no-store",
    });
  },

  /**
   * Fetch supported currencies and live exchange rates from the backend.
   */
  getCurrencies: async (): Promise<{ data: BackendCurrency[] }> => {
    return api.get<{ data: BackendCurrency[] }>("/api/v1/settings/currencies", {
      cache: "no-store",
    });
  },
};
