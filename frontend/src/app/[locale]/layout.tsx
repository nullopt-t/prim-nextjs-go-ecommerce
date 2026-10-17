import type { Metadata } from "next";
import { NextIntlClientProvider } from "next-intl";
import { getMessages, getTranslations } from "next-intl/server";
import { notFound } from "next/navigation";
import { routing } from "@/i18n/routing";
import { ThemeProvider } from "@/context/theme";
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

import { CartProvider } from "@/context/CartContext";
import { CatalogProvider } from "@/context/CatalogContext";
import { AuthProvider } from "@/context/AuthContext";
import { WishlistProvider } from "@/context/WishlistContext";
import { NotificationProvider } from "@/context/NotificationContext";

import { CartSidebar } from "@/components/ui/cartSidebar";
import { ResponsiveToaster } from "@/components/ui/responsiveToaster";

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

  const messages = await getMessages();
  const dir = locale === "ar" ? "rtl" : "ltr";

  return (
    <html
      lang={locale}
      dir={dir}
      className={`${ibmPlexSans.variable} ${ibmPlexSansArabic.variable}`}
      suppressHydrationWarning
    >
      <meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover" />
      <body className={`antialiased ${locale === "ar" ? "font-arabic" : "font-sans"}`}>
        <NextIntlClientProvider locale={locale} messages={messages}>
          <ThemeProvider>
            <AuthProvider>
              <WishlistProvider>
                <NotificationProvider>
                  <CatalogProvider>
                    <CartProvider>
                      {children}
                      <CartSidebar />
                    </CartProvider>
                  </CatalogProvider>
                </NotificationProvider>
              </WishlistProvider>
            </AuthProvider>
          </ThemeProvider>
          <ResponsiveToaster richColors />
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
