"use client";

import { Mail, Phone, MapPin } from "lucide-react";
import { usePublicSettings } from "@/context/PublicSettingsContext";

/**
 * Renders the store's contact identity from the admin settings. Falls back to
 * rendering nothing until settings load (or if the request fails).
 */
export function StoreContactInfo() {
  const { settings } = usePublicSettings();

  if (!settings) return null;

  const { storeEmail, storePhone, storeAddress } = settings;
  if (!storeEmail && !storePhone && !storeAddress) return null;

  return (
    <ul className="mt-4 flex flex-col gap-2 text-sm text-muted-foreground">
      {storeEmail && (
        <li className="flex items-center gap-2">
          <Mail className="size-3.5 text-accent-brand shrink-0" />
          <a href={`mailto:${storeEmail}`} className="hover:text-foreground transition-colors">
            {storeEmail}
          </a>
        </li>
      )}
      {storePhone && (
        <li className="flex items-center gap-2">
          <Phone className="size-3.5 text-accent-brand shrink-0" />
          <a href={`tel:${storePhone.replace(/\s+/g, "")}`} className="hover:text-foreground transition-colors" dir="ltr">
            {storePhone}
          </a>
        </li>
      )}
      {storeAddress && (
        <li className="flex items-start gap-2">
          <MapPin className="size-3.5 text-accent-brand shrink-0 mt-0.5" />
          <span>{storeAddress}</span>
        </li>
      )}
    </ul>
  );
}
