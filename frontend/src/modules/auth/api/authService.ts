import { api } from "@/shared/api/client";
import type { ApiResponse } from "@/shared/api/types";
import type {
  AuthChallengePayload,
  AuthVerifyPayload,
  Session,
  StartChallenge,
  User,
  VerifyChallengeResult,
} from "@/modules/auth/types";
import type { PasskeyAssertion } from "@/modules/auth/lib/passkey";

export type StartChallengePayload = AuthChallengePayload;
export type VerifyChallengePayload = AuthVerifyPayload;

export const authService = {
  /** Server-generated WebAuthn authentication options (passkeys when enabled). */
  getPasskeyAuthenticationOptions: () =>
    api.get<ApiResponse<{ challenge: string; rpId?: string; requestId?: string; userVerification?: string }>>(
      "/api/v1/auth/passkey/authentication/options",
      { skipAuthRedirect: true },
    ),

  /** Verifies a passkey assertion; sets the session cookie on success. */
  verifyPasskeyAssertion: (
    requestId: string | undefined,
    assertion: PasskeyAssertion,
  ) => {
    const response = assertion.response as AuthenticatorAssertionResponse;
    // Chunked: spreading a large Uint8Array into fromCharCode overflows the stack.
    const base64Url = (buffer: ArrayBuffer) => {
      const bytes = new Uint8Array(buffer);
      let binary = "";
      for (let i = 0; i < bytes.length; i += 0x8000) {
        binary += String.fromCharCode(...bytes.subarray(i, i + 0x8000));
      }
      return btoa(binary)
        .replace(/\+/g, "-")
        .replace(/\//g, "_")
        .replace(/=+$/, "");
    };
    return api.post<ApiResponse<User>>("/api/v1/auth/passkey/authentication/verify", {
      requestId,
      credential: {
        id: assertion.id,
        type: assertion.type,
        rawId: base64Url(assertion.rawId),
        response: {
          authenticatorData: base64Url(response.authenticatorData),
          clientDataJSON: base64Url(response.clientDataJSON),
          signature: base64Url(response.signature),
          userHandle: response.userHandle ? base64Url(response.userHandle) : null,
        },
      },
    });
  },

  startChallenge: (payload: AuthChallengePayload) =>
    api.post<ApiResponse<StartChallenge>, AuthChallengePayload>("/api/v1/auth/challenge/start", payload),

  verifyChallenge: (payload: AuthVerifyPayload) =>
    api.post<ApiResponse<VerifyChallengeResult>, AuthVerifyPayload>("/api/v1/auth/challenge/verify", payload),

  resendChallenge: (payload: AuthChallengePayload) =>
    api.post<ApiResponse<StartChallenge>, AuthChallengePayload>("/api/v1/auth/challenge/resend", payload),

  getMe: () => api.get<ApiResponse<User>>("/api/v1/auth/me"),

  getSessions: () => api.get<ApiResponse<Session[]>>("/api/v1/auth/sessions"),

  logout: () => api.post<ApiResponse<{ message: string }>>("/api/v1/auth/logout"),

  deleteSession: (sessionId: string) =>
    api.delete<void>(`/api/v1/auth/sessions/${sessionId}`),
};
