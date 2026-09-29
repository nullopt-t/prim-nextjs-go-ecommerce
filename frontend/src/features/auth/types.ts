export interface ValidationResult {
  isValid: boolean;
  error: string | null;
}

export interface AuthChallengePayload {
  email: string;
}

export interface AuthVerifyPayload {
  email?: string;
  code: string;
}

export type AuthFormType = "login" | "verify" | "register";
