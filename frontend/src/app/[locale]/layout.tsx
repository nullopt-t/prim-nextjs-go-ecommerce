import type { Metadata } from "next";
import { NextIntlClientProvider } from "next-intl";
import { getTranslations } from "next-intl/server";
import { getRouteMessages } from "@/shared/i18n/messages";
import { notFound, redirect } from "next/navigation";
import { cookies, headers } from "next/headers";
import { routing } from "@/shared/i18n/routing";
import { ThemeProvider } from "@/shared/theme";
import { IBM_Plex_Sans, IBM_Plex_Sans_Arabic } from "next/font/google";
import "@/styles/index.css";

const ibmPlexSans = IBM_Plex_Sans({
  subsets: ["latin"],
  weight: ["300", "400", "500", "600", "700"],
  variable: "--font-ibm-plex-sans",
  display: "swap",
});

const ibmPlexSansArabic = IBM_Plex_Sans_Arabic({
  subsets: ["arabic"],
  weight: ["300", "400", "500", "600", "700"],
  variable: "--font-ibm-plex-sans-arabic",
  display: "swap",
});

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("meta.home");
  return {
    metadataBase: new URL(
      process.env.NEXT_PUBLIC_SITE_URL || "http://localhost:3000",
    ),
    title: t("title"),
    description: t("description"),
    icons: {
      icon: [
        { url: "/favicon.svg", type: "image/svg+xml" },
      ],
      shortcut: "/favicon.svg",
      apple: "/favicon.svg",
    },
  };
}

export function generateStaticParams() {
  return routing.locales.map((locale) => ({ locale }));
}

import { CartProvider } from "@/modules/cart/model/cartContext";
import { AuthProvider } from "@/modules/auth/model/authContext";
import { WishlistProvider } from "@/modules/wishlist/model/wishlistContext";
import { NotificationProvider } from "@/modules/notifications/model/notificationContext";
import { PublicSettingsProvider } from "@/modules/settings/model/publicSettingsContext";
import { getPublicSettings } from "@/shared/lib/server/publicSettings";
import { EXPLICIT_LOCALE_COOKIE_NAME } from "@/shared/lib/localePreference";
import {
	resolveTargetLocale,
	shouldAddLocalePrefix,
	type SupportedLocale,
} from "@/shared/lib/server/locale";

import { CartSidebar } from "@/shared/ui/cartSidebar";
import { ResponsiveToaster } from "@/shared/ui/responsiveToaster";

export default async function RootLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;

  if (!routing.locales.includes(locale as any)) {
    notFound();
  }

  const [messages, publicSettings] = await Promise.all([
    getRouteMessages(),
    getPublicSettings(),
  ]);

  // Locale resolution is server-first:
  // 1. An explicit visitor choice stored in the short `locale` cookie wins.
  // 2. Otherwise the store's configured default language applies.
  const explicitLocale = (await cookies()).get(
    EXPLICIT_LOCALE_COOKIE_NAME,
  )?.value;
  const askedPath = (await headers()).get("x-pathname") || "/";
  const storeDefault = publicSettings?.defaultLocale;
  const targetLocale = resolveTargetLocale(explicitLocale, storeDefault);

  if (
  	targetLocale &&
  	shouldAddLocalePrefix(locale as SupportedLocale, askedPath, targetLocale)
  ) {
    const pathWithoutLocale =
      askedPath.replace(/^\/(en|ar)(?=\/|$)/, "") || "/";
    redirect(
      `/${targetLocale}${pathWithoutLocale === "/" ? "" : pathWithoutLocale}`,
    );
  }

  const dir = locale === "ar" ? "rtl" : "ltr";

  return (
    <html
      lang={locale}
      dir={dir}
      className={`${ibmPlexSans.variable} ${ibmPlexSansArabic.variable}`}
      suppressHydrationWarning
    >
      <head>
        <script
          // Apply the stored theme before first paint to avoid a light flash.
          dangerouslySetInnerHTML={{
            __html:
              "try{var t=localStorage.getItem('theme');if(t==='dark'||t==='light'){document.documentElement.setAttribute('data-theme',t)}}catch(e){}",
          }}
        />
      </head>
      <meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover" />
      <body className={`antialiased ${locale === "ar" ? "font-arabic" : "font-sans"}`}>
        <NextIntlClientProvider locale={locale} messages={messages}>
          <PublicSettingsProvider initialSettings={publicSettings}>
            <ThemeProvider>
              <AuthProvider>
                <WishlistProvider>
                  <NotificationProvider>
                    <CartProvider>
                      {children}
                      <CartSidebar />
                    </CartProvider>
                    <ResponsiveToaster richColors />
                  </NotificationProvider>
                </WishlistProvider>
              </AuthProvider>
            </ThemeProvider>
          </PublicSettingsProvider>
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
