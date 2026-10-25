import { ValidationResult } from "../types";

const OTP_CODE_REGEX = /^\d{6}$/;

export function validateOtpCode(code: string): ValidationResult {
  const trimmed = code.trim();
  if (!trimmed) {
    return { isValid: false, error: "Verification code is required." };
  }
  if (!OTP_CODE_REGEX.test(trimmed)) {
    return { isValid: false, error: "Verification code must contain 6 digits." };
  }
  return { isValid: true, error: null };
}
