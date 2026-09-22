import { api } from "@/api/client";

export const authService = {
	startChallenge: (payload) => api.post("/api/v1/auth/challenge/start", payload),
	verifyChallenge: (payload) => api.post("/api/v1/auth/challenge/verify", payload),
	resendChallenge: (payload) => api.post("/api/v1/auth/challenge/resend", payload),
	getMe: () => api.get("/api/v1/auth/me"),
	getSessions: () => api.get("/api/v1/auth/sessions"),
	logout: (sessionId) => api.delete(`/api/v1/auth/sessions/${sessionId}`),
};
