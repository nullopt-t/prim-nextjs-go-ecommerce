import { api } from "@/shared/api/client";
import type { ApiResponse } from "@/shared/api/types";
import type {
  AuthChallengePayload,
  AuthVerifyPayload,
  Session,
  StartChallenge,
  User,
} from "@/modules/auth/types";

export type StartChallengePayload = AuthChallengePayload;
export type VerifyChallengePayload = AuthVerifyPayload;

export const authService = {
  startChallenge: (payload: AuthChallengePayload) =>
    api.post<ApiResponse<StartChallenge>, AuthChallengePayload>("/api/v1/auth/challenge/start", payload),

  verifyChallenge: (payload: AuthVerifyPayload) =>
    api.post<ApiResponse<User>, AuthVerifyPayload>("/api/v1/auth/challenge/verify", payload),

  resendChallenge: (payload: AuthChallengePayload) =>
    api.post<ApiResponse<StartChallenge>, AuthChallengePayload>("/api/v1/auth/challenge/resend", payload),

  getMe: () => api.get<ApiResponse<User>>("/api/v1/auth/me"),

  getSessions: () => api.get<ApiResponse<Session[]>>("/api/v1/auth/sessions"),

  logout: () => api.post<ApiResponse<{ message: string }>>("/api/v1/auth/logout"),

  deleteSession: (sessionId: string) =>
    api.delete<void>(`/api/v1/auth/sessions/${sessionId}`),
};
