import { Brand } from "@/shared/ui/Brand";
import { Link } from "@/shared/i18n/navigation";
import { getTranslations } from "next-intl/server";
import { RouteMessages } from "@/shared/i18n/messages";

export default async function MarketingLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const t = await getTranslations("marketing");

  return (
    <RouteMessages group="marketing">
    <div className="min-h-screen bg-background text-foreground">
      <header className="sticky top-0 z-40 border-b border-border bg-background/90 backdrop-blur-md">
        <div className="mx-auto flex h-15 max-w-6xl items-center justify-between gap-4 px-4 sm:px-6">
          <Link href="/" className="flex items-center gap-2">
            <Brand />
          </Link>
          <nav className="flex items-center gap-2">
            <Link
              href="/"
              className="hidden rounded-md px-3 py-2 text-sm font-medium text-muted-foreground hover:text-foreground sm:inline-flex"
            >
              {t("nav.storefront")}
            </Link>
            <Link
              href="/auth?redirect=%2Fseller"
              className="rounded-md px-3 py-2 text-sm font-medium text-muted-foreground hover:text-foreground"
            >
              {t("nav.signIn")}
            </Link>
            <Link
              href="/seller"
              className="rounded-md bg-accent-brand px-3.5 py-2 text-sm font-semibold text-white hover:bg-accent-brand/90"
            >
              {t("nav.startSelling")}
            </Link>
          </nav>
        </div>
      </header>

      {children}

      <footer className="border-t border-border">
        <div className="mx-auto flex max-w-6xl flex-col items-center justify-between gap-3 px-4 py-6 text-xs text-muted-foreground sm:flex-row sm:px-6">
          <span>{t("footer.copyright", { year: new Date().getFullYear() })}</span>
          <div className="flex items-center gap-4">
            <Link href="/" className="hover:text-foreground">
              {t("nav.storefront")}
            </Link>
            <Link href="/contact" className="hover:text-foreground">
              {t("footer.contact")}
            </Link>
            <a href="mailto:sellers@marketplace.com" className="hover:text-foreground">
              {t("footer.merchantSupport")}
            </a>
          </div>
        </div>
      </footer>
    </div>
    </RouteMessages>
  );
}
