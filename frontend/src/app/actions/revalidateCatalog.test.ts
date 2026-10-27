import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";

vi.mock("next/cache", () => ({
	updateTag: vi.fn(),
}));
vi.mock("next/headers", () => ({
	cookies: vi.fn(),
}));

import { updateTag } from "next/cache";
import { cookies, type RequestCookies } from "next/headers";
import { revalidateCatalog } from "@/app/actions/revalidateCatalog";

const fetchMock = vi.fn();

const mockCookies = vi.mocked(cookies);
const mockUpdateTag = vi.mocked(updateTag);

beforeEach(() => {
	process.env.NEXT_PUBLIC_API_URL = "http://api:8081";
	vi.resetAllMocks();
	fetchMock.mockReset();
	(globalThis as unknown as { fetch?: typeof fetch }).fetch = fetchMock;
});

afterEach(() => {
	delete process.env.NEXT_PUBLIC_API_URL;
});

describe("revalidateCatalog", () => {
	it("returns false when the API URL is not configured", async () => {
		delete process.env.NEXT_PUBLIC_API_URL;

		expect(await revalidateCatalog()).toBe(false);
		expect(fetchMock).not.toHaveBeenCalled();
		expect(mockUpdateTag).not.toHaveBeenCalled();
	});

	it("returns false when there are no session cookies", async () => {
		vi.mocked(cookies).mockResolvedValue({ getAll: () => [] } as unknown as RequestCookies);

		expect(await revalidateCatalog()).toBe(false);
		expect(fetchMock).not.toHaveBeenCalled();
		expect(mockUpdateTag).not.toHaveBeenCalled();
	});

	it("returns false when the session check fails", async () => {
		vi.mocked(cookies).mockResolvedValue(
			{ getAll: () => [{ name: "session", value: "abc" }] } as unknown as RequestCookies
		);
		fetchMock.mockResolvedValueOnce({ ok: false, status: 401 } as Response);

		expect(await revalidateCatalog()).toBe(false);
		expect(fetchMock).toHaveBeenCalledTimes(1);
		expect(mockUpdateTag).not.toHaveBeenCalled();
	});

	it("returns false when the caller is not an admin", async () => {
		vi.mocked(cookies).mockResolvedValue(
			{ getAll: () => [{ name: "session", value: "abc" }] } as unknown as RequestCookies
		);
		fetchMock.mockResolvedValueOnce({
			ok: true,
			json: async () => ({ data: { role: "customer" } }),
		} as Response);

		expect(await revalidateCatalog()).toBe(false);
		expect(mockUpdateTag).not.toHaveBeenCalled();
	});

	it("revalidates all catalog tags and returns true for an admin", async () => {
		vi.mocked(cookies).mockResolvedValue(
			{ getAll: () => [{ name: "session", value: "abc" }] } as unknown as RequestCookies
		);
		fetchMock.mockResolvedValueOnce({
			ok: true,
			json: async () => ({ data: { role: "superadmin" } }),
		} as Response);

		expect(await revalidateCatalog()).toBe(true);

		expect(fetchMock).toHaveBeenCalledWith(
			"http://api:8081/api/v1/auth/me",
			expect.objectContaining({
				headers: expect.objectContaining({ Cookie: "session=abc" }),
			})
		);
		expect(mockUpdateTag).toHaveBeenCalledWith("catalog:products");
		expect(mockUpdateTag).toHaveBeenCalledWith("catalog:categories");
		expect(mockUpdateTag).toHaveBeenCalledWith("catalog:brands");
		expect(mockUpdateTag).toHaveBeenCalledTimes(3);
	});
});
