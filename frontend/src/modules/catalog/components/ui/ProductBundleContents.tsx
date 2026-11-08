import { useEffect, useMemo, useState } from "react";
import { Link } from "@/shared/i18n/navigation";
import { LazyImage } from "@/shared/ui/lazyImage";
import { BadgePercent, CheckCircle2, XCircle, ShoppingBag } from "lucide-react";
import { useTranslations } from "next-intl";
import { catalogService } from "@/modules/catalog/api/catalogService";
import { normalizeProduct } from "@/modules/catalog/model/normalizeProduct";
import { computeBundleSavings, resolveItemPrice } from "@/modules/catalog/utils/bundlePricing";
import { usePriceFormatter } from "@/modules/settings/model/usePriceFormatter";
import type { ProductBundle } from "../../types";

interface ProductBundleContentsProps {
  bundle?: ProductBundle | null;
  /** Price of the kit itself (the bundle product's own price). */
  bundlePrice?: number;
  /** Adds the kit to the cart; the kit sells as a single product. */
  onBuyBundle?: () => void;
}

/**
 * Lists what a kit ships with and the availability of each component. The
 * payload carries no prices, so savings are computed from each component
 * product's own price (resolved client-side); when a component can't be
 * priced, the savings line simply hides.
 */
export function ProductBundleContents({ bundle, bundlePrice, onBuyBundle }: ProductBundleContentsProps) {
  const t = useTranslations("products.bundle");
  const formatPrice = usePriceFormatter();
  const [prices, setPrices] = useState<Record<string, number>>({});

  const slugsKey = useMemo(
    () => (bundle ? bundle.items.map((item) => item.slug).join("|") : ""),
    [bundle],
  );

  useEffect(() => {
    if (!bundle || bundle.items.length === 0) return;
    let isMounted = true;
    const items = bundle.items;
    Promise.allSettled(
      items.map(async (item) => {
        const res = await catalogService.getProductBySlug(item.slug);
        const raw = (res as { data?: unknown })?.data ?? (res as unknown);
        const product = normalizeProduct(raw as Parameters<typeof normalizeProduct>[0]) as unknown as {
          price?: string | number;
          extractedPrice?: number;
          variants?: Array<{ id?: string; sku?: string; extractedPrice?: number; price?: string | number; isDefault?: boolean }>;
        } | null;
        return { slug: item.slug, price: product ? resolveItemPrice(product, item.variant) : 0 };
      }),
    ).then((results) => {
      if (!isMounted) return;
      const next: Record<string, number> = {};
      results.forEach((result, index) => {
        if (result.status === "fulfilled" && result.value.price > 0) {
          next[items[index].slug] = result.value.price;
        }
      });
      setPrices(next);
    });
    return () => {
      isMounted = false;
    };
  }, [bundle, slugsKey]);

  if (!bundle || bundle.items.length === 0) return null;

  const savings = computeBundleSavings(
    bundle.items.map((item) => ({
      slug: item.slug,
      quantity: item.quantity,
      price: prices[item.slug] ?? 0,
    })),
    bundlePrice ?? 0,
  );

  return (
    <section className="mt-8 rounded-2xl border border-border bg-card p-4 shadow-sm sm:mt-10 sm:p-7">
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

      {savings && (
        <p className="mt-4 inline-flex items-center gap-2 rounded-full bg-emerald-500/10 px-3.5 py-1.5 text-xs font-semibold text-emerald-600 dark:text-emerald-400 sm:text-sm">
          <BadgePercent className="size-4" aria-hidden="true" />
          {t("savings", {
            amount: formatPrice(savings.savings),
            percent: savings.percentOff,
          })}
        </p>
      )}

      <ul className="mt-5 grid grid-cols-1 gap-3 sm:grid-cols-2">
        {bundle.items.map((item) => {
          const linePrice = prices[item.slug];
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

                {linePrice !== undefined && (
                  <p className="mt-0.5 text-xs font-semibold text-foreground">
                    {formatPrice(linePrice * item.quantity)}
                  </p>
                )}

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

      {onBuyBundle && (
        <div className="mt-5 flex justify-end border-t border-border/70 pt-5">
          <button
            type="button"
            onClick={onBuyBundle}
            disabled={!bundle.isInStock}
            className="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-accent-brand px-5 py-3 text-sm font-semibold text-white transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50 sm:w-auto"
          >
            <ShoppingBag className="size-4" aria-hidden="true" />
            {t("buyBundle", {
              price:
                bundlePrice !== undefined
                  ? formatPrice(bundlePrice)
                  : "",
            })}
          </button>
        </div>
      )}
    </section>
  );
}
