import type { PublicSettings } from "@/modules/settings/api/settingsService";
import { getServerApiBaseUrl } from "@/shared/lib/server/apiBaseUrl";

export const PUBLIC_SETTINGS_TAG = "public-settings";

/**
 * Server-side fetch for public storefront settings. Used to render settings
 * dependent UI (brand, footer, announcement) in the initial HTML response.
 */
export async function getPublicSettings(): Promise<PublicSettings | null> {
  const baseUrl = getServerApiBaseUrl();
  if (!baseUrl) return null;

  try {
    const response = await fetch(
      `${baseUrl}/api/v1/settings/public`,
      {
        next: { revalidate: 60, tags: [PUBLIC_SETTINGS_TAG] },
        headers: { Accept: "application/json" },
      },
    );
    if (!response.ok) return null;

    const payload = await response.json();
    return payload?.data ?? null;
  } catch {
    // Settings are non-critical for rendering; the client provider will retry.
    return null;
  }
}
