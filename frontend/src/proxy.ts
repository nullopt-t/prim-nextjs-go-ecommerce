import createMiddleware from "next-intl/middleware";
import type { NextRequest } from "next/server";
import { routing } from "@/shared/i18n/routing";

const intlMiddleware = createMiddleware(routing);

export default function proxy(request: NextRequest) {
	// Expose the requested path to Server Components so the locale layout can
	// redirect to the store's configured default language without a client
	// effect. The locale prefix is preserved in the header so the layout can
	// tell an explicit /ar or /en URL from an unprefixed one.
	request.headers.set(
		"x-pathname",
		`${request.nextUrl.pathname}${request.nextUrl.search}`,
	);

	return intlMiddleware(request);
}

export const config = {
	// Match all pathnames except for
	// - … if they start with `/api`, `/_next` or `/_vercel`
	// - … the ones containing a dot (e.g. `favicon.ico`)
	matcher: ["/((?!api|_next|_vercel|.*\\..*).*)"],
};
