const AuthEndpoints = {
	// fetch session user data
	me: "/auth/me",
	// Starts an authentication challenge by sending a verification code to the provided email or phone number.
	start: "/auth/challenge/start",
	// Verifies the one-time code sent to the user's email or phone and returns an access token and refresh token.
	verifyCode: "/auth/challenge/verify",
	// Resends a new verification code to the provided email or phone number if allowed by the challenge policy.
	resendCode: "/auth/challenge/resend",
	// Rotate refresh token and issue new access and refresh tokens.
	refresh: "/auth/refresh",
};

export default AuthEndpoints;
