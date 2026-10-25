import { api } from "@/shared/api/client";

export interface StartChallengePayload {
	email: string;
}

export interface VerifyChallengePayload {
	email?: string;
	code: string;
	merge_cart?: boolean;
}

export const authService = {
	startChallenge: (payload: StartChallengePayload) =>
		api.post("/api/v1/auth/challenge/start", payload),
	verifyChallenge: (payload: VerifyChallengePayload) =>
		api.post("/api/v1/auth/challenge/verify", payload),
	resendChallenge: (payload: StartChallengePayload) =>
		api.post("/api/v1/auth/challenge/resend", payload),
	getMe: () => api.get("/api/v1/auth/me"),
	getSessions: () => api.get("/api/v1/auth/sessions"),
	logout: () => api.post("/api/v1/auth/logout"),
	deleteSession: (sessionId: string) => api.delete(`/api/v1/auth/sessions/${sessionId}`),
};
