import type { Metadata } from "next";
import { NextIntlClientProvider } from "next-intl";
import { getMessages } from "next-intl/server";
import { notFound } from "next/navigation";
import { routing } from "@/i18n/routing";
import { ThemeProvider } from "@/context/theme";
import "@/styles/index.css";

export const metadata: Metadata = {
  title: "PRIM - Premium Audio & Wearables",
  description: "E-Commerce store for audio gear, wearables, and accessories",
};

export function generateStaticParams() {
  return routing.locales.map((locale) => ({ locale }));
}

import { CartProvider } from "@/context/CartContext";
import { CatalogProvider } from "@/context/CatalogContext";
import { AuthProvider } from "@/context/AuthContext";
import { PortalSwitcher } from "@/components/ui/portalSwitcher";
import { CartSidebar } from "@/components/ui/cartSidebar";

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
    <html lang={locale} dir={dir} suppressHydrationWarning>
      <body className="antialiased">
        <NextIntlClientProvider locale={locale} messages={messages}>
          <ThemeProvider>
            <AuthProvider>
              <CatalogProvider>
                <CartProvider>
                  {children}
                  <CartSidebar />
                  <PortalSwitcher />
                </CartProvider>
              </CatalogProvider>
            </AuthProvider>
          </ThemeProvider>
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
