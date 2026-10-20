"use server";

import { updateTag } from "next/cache";
import { cookies } from "next/headers";
import { PUBLIC_SETTINGS_TAG } from "@/lib/server/publicSettings";

/**
 * Invalidate the server cache for public settings after an admin save.
 * The backend call verifies that the caller still has an admin session.
 */
export async function revalidatePublicSettings(): Promise<boolean> {
  const apiUrl = process.env.API_URL || process.env.NEXT_PUBLIC_API_URL;
  if (!apiUrl) return false;

  const cookieStore = await cookies();
  const cookieHeader = cookieStore
    .getAll()
    .map((cookie) => `${cookie.name}=${cookie.value}`)
    .join("; ");
  if (!cookieHeader) return false;

  const response = await fetch(`${apiUrl.replace(/\/$/, "")}/api/v1/admin/settings`, {
    cache: "no-store",
    headers: {
      Accept: "application/json",
      Cookie: cookieHeader,
    },
  });
  if (!response.ok) return false;

  updateTag(PUBLIC_SETTINGS_TAG);
  return true;
}
