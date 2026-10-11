export const AuthEndpoints = {
  // fetch session user data
  me: "/api/v1/auth/me",
  // Starts an authentication challenge by sending a verification code to the provided email or phone number.
  start: "/api/v1/auth/challenge/start",
  // Verifies the one-time code sent to the user's email or phone and establishes an authenticated session cookie.
  verifyCode: "/api/v1/auth/challenge/verify",
  // Resends a new verification code to the provided email or phone number if allowed by the challenge policy.
  resendCode: "/api/v1/auth/challenge/resend",
  // Active sessions management
  sessions: "/api/v1/auth/sessions",
  logout: "/api/v1/auth/logout",
} as const;

export default AuthEndpoints;
