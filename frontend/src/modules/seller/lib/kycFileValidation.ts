export const KYC_FILE_EXTENSIONS = [".pdf", ".jpg", ".jpeg", ".png", ".webp"];
export const KYC_MAX_FILE_SIZE = 10 * 1024 * 1024;

export type KycFileError = "invalidType" | "emptyFile" | "tooLarge";

export function validateKycFile(file: File): KycFileError | null {
	if (!KYC_FILE_EXTENSIONS.some((ext) => file.name.toLowerCase().endsWith(ext))) {
		return "invalidType";
	}
	if (file.size === 0) return "emptyFile";
	if (file.size > KYC_MAX_FILE_SIZE) return "tooLarge";
	return null;
}
