import { Link } from "@/shared/i18n/navigation";
import { LazyImage } from "@/shared/ui/lazyImage";
import { CheckCircle2, XCircle } from "lucide-react";
import { useTranslations } from "next-intl";
import type { ProductBundle } from "../../types";

interface ProductBundleContentsProps {
  bundle?: ProductBundle | null;
}

/**
 * Lists what a kit ships with and the availability of each component. A bundle
 * keeps no stock of its own, so this section is the only place a customer can
 * see why a kit is or is not available.
 */
export function ProductBundleContents({ bundle }: ProductBundleContentsProps) {
  const t = useTranslations("products.bundle");

  if (!bundle || bundle.items.length === 0) return null;

  return (
    <section className="mt-8 rounded-2xl border border-border bg-card p-5 shadow-sm sm:mt-10 sm:p-7">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-bold text-foreground sm:text-xl">{t("title")}</h2>
          <p className="mt-1 text-xs text-muted-foreground sm:text-sm">
            {t("items", { count: bundle.itemCount })}
          </p>
        </div>

        <span
          className={`inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-semibold ${
            bundle.isInStock
              ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
              : "bg-destructive/10 text-destructive"
          }`}
        >
          {bundle.isInStock ? (
            <CheckCircle2 className="size-3.5" aria-hidden="true" />
          ) : (
            <XCircle className="size-3.5" aria-hidden="true" />
          )}
          {bundle.isInStock
            ? t("kitAvailable", { count: bundle.availableQuantity })
            : t("kitUnavailable")}
        </span>
      </div>

      <ul className="mt-5 grid grid-cols-1 gap-3 sm:grid-cols-2">
        {bundle.items.map((item) => {
          const content = (
            <>
              <LazyImage
                src={item.thumbnail}
                alt={item.title}
                containerClassName="size-16 shrink-0 rounded-xl border border-border"
                imageClassName="size-full object-cover"
              />

              <div className="min-w-0 flex-1">
                <div className="flex items-start justify-between gap-2">
                  <span dir="auto" className="text-sm font-medium text-foreground line-clamp-2">
                    {item.title}
                  </span>
                  <span className="shrink-0 rounded-md bg-muted px-1.5 py-0.5 text-[11px] font-semibold text-muted-foreground">
                    {t("quantity", { count: item.quantity })}
                  </span>
                </div>

                {item.variant ? (
                  <p dir="auto" className="mt-0.5 truncate text-xs text-muted-foreground">
                    {item.variant}
                  </p>
                ) : null}

                <p
                  className={`mt-1 flex items-center gap-1 text-xs font-medium ${
                    item.isInStock ? "text-emerald-600 dark:text-emerald-400" : "text-destructive"
                  }`}
                >
                  {item.isInStock ? (
                    <CheckCircle2 className="size-3.5" aria-hidden="true" />
                  ) : (
                    <XCircle className="size-3.5" aria-hidden="true" />
                  )}
                  {item.isInStock ? t("itemInStock") : t("itemOutOfStock")}
                </p>
              </div>
            </>
          );

          return (
            <li key={`${item.id}-${item.variant ?? ""}`}>
              {item.slug ? (
                <Link
                  href={`/products/${item.slug}`}
                  className="flex gap-3 rounded-xl border border-border/70 p-3 transition-colors hover:border-accent-brand/50 hover:bg-muted/40"
                >
                  {content}
                </Link>
              ) : (
                <div className="flex gap-3 rounded-xl border border-border/70 p-3">{content}</div>
              )}
            </li>
          );
        })}
      </ul>

      <p className="mt-4 text-xs text-muted-foreground">{t("hint")}</p>
    </section>
  );
}
