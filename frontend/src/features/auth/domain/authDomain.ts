/**
 * Domain-level validation functions for Authentication (free of UI/DOM dependencies)
 */

export interface ValidationResult {
  isValid: boolean;
  error: string | null;
}

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const OTP_CODE_REGEX = /^\d{6}$/;

export function validateEmail(email: string): ValidationResult {
  const trimmed = email.trim();
  if (!trimmed) {
    return { isValid: false, error: "Email is required." };
  }
  if (!EMAIL_REGEX.test(trimmed)) {
    return { isValid: false, error: "Please enter a valid email address." };
  }
  return { isValid: true, error: null };
}

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
