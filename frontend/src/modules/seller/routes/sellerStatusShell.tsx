"use client";

import { Clock3, ShieldAlert, XCircle } from "lucide-react";
import { useTranslations } from "next-intl";
import { Link, useRouter } from "@/shared/i18n/navigation";
import { useAuthContext } from "@/modules/auth/model/authContext";
import { useSellerContext } from "./sellerContext";

const ICONS = {
  pending: Clock3,
  suspended: ShieldAlert,
  rejected: XCircle,
} as const;

export function SellerStatusShell({
  kind,
  action,
}: {
  kind: "pending" | "suspended" | "rejected";
  action?: React.ReactNode;
}) {
  const t = useTranslations("seller.statusPages");
  const router = useRouter();
  const { logout } = useAuthContext();
  const { seller } = useSellerContext();
  const Icon = ICONS[kind];

  const handleLogout = async () => {
    await logout();
    router.replace("/");
  };

  return (
    <div className="flex min-h-screen flex-col bg-muted/20 text-foreground">
      <header className="flex h-14 items-center justify-between border-b border-border bg-background px-4 md:px-6">
        <Link href="/" className="flex items-center gap-2">
          <span dir="ltr" className="text-lg font-black tracking-tighter">
            PRI<span className="text-accent-brand">M</span>
          </span>
          <span className="border-s border-border ps-3 text-xs font-medium text-muted-foreground">
            {t("sellerHub")}
          </span>
        </Link>
        <button
          type="button"
          onClick={handleLogout}
          className="text-xs font-medium text-muted-foreground hover:text-foreground hover:underline"
        >
          {t("signOut")}
        </button>
      </header>

      <main className="flex flex-1 items-center justify-center p-4 md:p-8">
        <section className="w-full max-w-xl rounded-xl border border-border bg-background p-6 shadow-sm md:p-8">
          <div className="flex size-11 items-center justify-center rounded-full bg-secondary">
            <Icon className="size-5 text-muted-foreground" />
          </div>
          <div className="mt-4 inline-flex rounded-full border border-border px-2.5 py-0.5 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
            {t(`kind.${kind}`)}
          </div>
          <h1 className="mt-3 text-xl font-semibold tracking-tight md:text-2xl">
            {t(`${kind}.title`)}
          </h1>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">
            {t(`${kind}.description`)}
          </p>

          {seller ? (
            <div className="mt-5 rounded-lg border border-border bg-muted/30 p-3">
              <div className="text-sm font-medium">{seller.storeName}</div>
              <div dir="ltr" className="mt-0.5 text-xs text-muted-foreground">
                /store/{seller.slug}
              </div>
              <div className="mt-2 text-[11px] text-muted-foreground">
                {t("applicationId")}: <span className="font-mono">{seller.id}</span>
              </div>
            </div>
          ) : null}

          {kind === "suspended" || kind === "rejected" ? (
            <div className="mt-4 rounded-lg border border-border p-3">
              <div className="text-xs font-medium text-muted-foreground">
                {t("reasonLabel")}
              </div>
              <p className="mt-1 text-sm">
                {seller?.statusReason || t("reasonFallback")}
              </p>
            </div>
          ) : null}

          {kind === "rejected" ? (
            <p className="mt-4 text-xs leading-5 text-muted-foreground">
              {t("rejected.note")}
            </p>
          ) : null}

          <div className="mt-6 flex flex-wrap items-center gap-2">
            {action}
            <a
              href="mailto:sellers@marketplace.com"
              className="inline-flex h-9 items-center rounded-md border border-border px-3 text-sm font-medium hover:bg-secondary"
            >
              {t("contactSupport")}
            </a>
            <Link
              href="/"
              className="inline-flex h-9 items-center rounded-md px-3 text-sm font-medium text-muted-foreground hover:text-foreground"
            >
              {t("storefront")}
            </Link>
          </div>
        </section>
      </main>
    </div>
  );
}
