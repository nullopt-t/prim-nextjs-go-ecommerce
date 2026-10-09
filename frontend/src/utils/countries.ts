// ISO 3166-1 alpha-2 country codes
export const COUNTRY_CODES = [
  "EG", // Egypt (Primary)
  "SA", // Saudi Arabia
  "AE", // UAE
  "KW", // Kuwait
  "QA", // Qatar
  "BH", // Bahrain
  "OM", // Oman
  "JO", // Jordan
  "LB", // Lebanon
  "IQ", // Iraq
  "US", // United States
  "GB", // United Kingdom
  "CA", // Canada
  "DE", // Germany
  "FR", // France
  "IT", // Italy
  "ES", // Spain
  "TR", // Turkey
  "MA", // Morocco
  "TN", // Tunisia
  "DZ", // Algeria
  "LY", // Libya
  "SD", // Sudan
  "YE", // Yemen
  "SY", // Syria
  "PS", // Palestine
  "NL", // Netherlands
  "BE", // Belgium
  "CH", // Switzerland
  "SE", // Sweden
  "NO", // Norway
  "DK", // Denmark
  "JP", // Japan
  "KR", // South Korea
  "CN", // China
  "IN", // India
  "AU", // Australia
  "NZ", // New Zealand
  "BR", // Brazil
  "MX", // Mexico
  "ZA", // South Africa
] as const;

export interface CountryOption {
  code: string;
  name: string;
  nativeName: string;
}

export function getLocalizedCountries(locale: string = "en"): CountryOption[] {
  let displayNames: Intl.DisplayNames | null = null;
  let nativeDisplayNames: Intl.DisplayNames | null = null;

  try {
    displayNames = new Intl.DisplayNames([locale], { type: "region" });
    nativeDisplayNames = new Intl.DisplayNames(["en"], { type: "region" });
  } catch {
    // Fallback if Intl.DisplayNames is not supported
  }

  const countries = COUNTRY_CODES.map((code) => {
    const localized = displayNames?.of(code) || code;
    const english = nativeDisplayNames?.of(code) || code;
    return {
      code,
      name: localized,
      nativeName: english,
    };
  });

  return countries.sort((a, b) => a.name.localeCompare(b.name, locale));
}
