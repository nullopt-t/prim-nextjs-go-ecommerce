import { describe, expect, it } from "vitest";
import { KYC_MAX_FILE_SIZE, validateKycFile } from "./kycFileValidation";

function kycFile(name: string, size: number): File {
	return { name, size } as File;
}

describe("validateKycFile", () => {
	it("accepts allowed types with content within the size limit", () => {
		expect(validateKycFile(kycFile("register.PDF", 1024))).toBeNull();
		expect(validateKycFile(kycFile("id.jpeg", 1024))).toBeNull();
		expect(validateKycFile(kycFile("id.WEBP", KYC_MAX_FILE_SIZE))).toBeNull();
	});

	it("rejects disallowed extensions", () => {
		expect(validateKycFile(kycFile("malware.exe", 1024))).toBe("invalidType");
		expect(validateKycFile(kycFile("id.png.exe", 1024))).toBe("invalidType");
	});

	it("rejects empty files", () => {
		expect(validateKycFile(kycFile("id.png", 0))).toBe("emptyFile");
	});

	it("rejects files above 10 MB", () => {
		expect(validateKycFile(kycFile("id.png", KYC_MAX_FILE_SIZE + 1))).toBe("tooLarge");
	});
});
