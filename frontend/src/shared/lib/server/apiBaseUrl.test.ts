import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { getServerApiBaseUrl } from "@/shared/lib/server/apiBaseUrl";

const originalServerUrl = process.env.API_URL;
const originalPublicUrl = process.env.NEXT_PUBLIC_API_URL;

beforeEach(() => {
	delete process.env.API_URL;
	delete process.env.NEXT_PUBLIC_API_URL;
});

afterEach(() => {
	if (originalServerUrl === undefined) delete process.env.API_URL;
	else process.env.API_URL = originalServerUrl;

	if (originalPublicUrl === undefined) delete process.env.NEXT_PUBLIC_API_URL;
	else process.env.NEXT_PUBLIC_API_URL = originalPublicUrl;
});

describe("getServerApiBaseUrl", () => {
	it("returns null when neither variable is configured", () => {
		expect(getServerApiBaseUrl()).toBeNull();
	});

	it("falls back to the public URL", () => {
		process.env.NEXT_PUBLIC_API_URL = "http://192.168.1.16:8081/";
		expect(getServerApiBaseUrl()).toBe("http://192.168.1.16:8081");
	});

	it("prefers the server-only API_URL over the public URL", () => {
		process.env.API_URL = "http://api:8081/";
		process.env.NEXT_PUBLIC_API_URL = "http://192.168.1.16:8081";
		expect(getServerApiBaseUrl()).toBe("http://api:8081");
	});
});
