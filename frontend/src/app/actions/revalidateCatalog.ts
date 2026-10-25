"use server";

import { updateTag } from "next/cache";
import { cookies } from "next/headers";
import {
	BRANDS_TAG,
	CATEGORIES_TAG,
	PRODUCTS_TAG,
} from "@/shared/lib/server/catalog";

/**
 * Invalidate the cached storefront catalog after an admin mutation.
 *
 * Server Actions are publicly callable, so verify the caller's session still
 * has an admin role before dropping the cache.
 */
export async function revalidateCatalog(): Promise<boolean> {
	const apiUrl = process.env.API_URL || process.env.NEXT_PUBLIC_API_URL;
	if (!apiUrl) return false;

	const cookieStore = await cookies();
	const cookieHeader = cookieStore
		.getAll()
		.map((cookie) => `${cookie.name}=${cookie.value}`)
		.join("; ");
	if (!cookieHeader) return false;

	try {
		const response = await fetch(`${apiUrl.replace(/\/$/, "")}/api/v1/auth/me`, {
			cache: "no-store",
			headers: { Accept: "application/json", Cookie: cookieHeader },
		});
		if (!response.ok) return false;
		const payload = await response.json();
		const user = payload?.data ?? payload;
		if (!["admin", "superadmin"].includes(user?.role)) return false;
	} catch {
		return false;
	}

	updateTag(PRODUCTS_TAG);
	updateTag(CATEGORIES_TAG);
	updateTag(BRANDS_TAG);
	return true;
}
