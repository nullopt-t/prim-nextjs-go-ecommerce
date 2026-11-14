/**
 * Auth domain and response models matching backend auth handlers.
 */

export interface User {
  id: string;
  email?: string;
  identifier?: string;
  name?: string;
  role?: string;
  status?: string;
  phone?: string;
  avatar?: string;
  createdAt?: string;
}

export interface StartChallenge {
  email: string;
  expiresAt: string;
  duration: number;
}

export interface Session {
  id: string;
  userId?: string;
  userAgent?: string;
  ipAddress?: string;
  createdAt: string;
  lastActive?: string;
  current?: boolean;
}

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
  fullName?: string;
  merge_cart?: boolean;
  mergeCart?: boolean;
}

export interface VerifyChallengeResult {
  profileRequired?: boolean;
}

export type AuthFormType = "login" | "verify" | "register";
