"use server";

import { updateTag } from "next/cache";
import { cookies } from "next/headers";
import { PRODUCTS_TAG } from "@/shared/lib/server/catalog";
import { getServerApiBaseUrl } from "@/shared/lib/server/apiBaseUrl";

/**
 * Invalidate the cached storefront product list after a seller mutation.
 *
 * Server Actions are publicly callable, so verify the caller's session
 * still belongs to an active seller before dropping the cache.
 */
export async function revalidateCatalogForSeller(): Promise<boolean> {
	const apiUrl = getServerApiBaseUrl();
	if (!apiUrl) return false;

	const cookieStore = await cookies();
	const cookieHeader = cookieStore
		.getAll()
		.map((cookie) => `${cookie.name}=${cookie.value}`)
		.join("; ");
	if (!cookieHeader) return false;

	try {
		const response = await fetch(`${apiUrl}/api/v1/seller/profile`, {
			cache: "no-store",
			headers: { Accept: "application/json", Cookie: cookieHeader },
		});
		if (!response.ok) return false;
	} catch {
		return false;
	}

	updateTag(PRODUCTS_TAG);
	return true;
}
