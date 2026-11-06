export const MIN_SLUG_LENGTH = 3;
export const MAX_SLUG_LENGTH = 63;

const SLUG_PATTERN = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

export const RESERVED_SLUGS = new Set([
  "admin",
  "api",
  "auth",
  "seller",
  "sell",
  "store",
  "stores",
  "products",
  "categories",
  "brands",
  "cart",
  "checkout",
  "account",
  "user",
  "orders",
  "payments",
  "settings",
  "support",
  "help",
  "about",
  "terms",
  "privacy",
  "login",
  "register",
  "search",
  "wishlist",
  "en",
  "ar",
]);

export interface SlugValidation {
  slug: string;
  reason?: string;
}

export function normalizeSlug(raw: string): string {
  return raw.normalize("NFKC").trim().toLowerCase();
}

export function validateSlug(raw: string): SlugValidation {
  const slug = normalizeSlug(raw);
  if (!slug) return { slug, reason: "Slug is required" };
  if (slug.length < MIN_SLUG_LENGTH || slug.length > MAX_SLUG_LENGTH) {
    return { slug, reason: "Slug must be between 3 and 63 characters" };
  }
  if (!SLUG_PATTERN.test(slug)) {
    return {
      slug,
      reason: "Use lowercase letters, numbers, and single hyphens only",
    };
  }
  if (RESERVED_SLUGS.has(slug)) {
    return { slug, reason: "This slug is a reserved platform keyword" };
  }
  return { slug };
}
