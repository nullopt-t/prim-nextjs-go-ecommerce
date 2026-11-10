import { useEffect, useRef, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import {
  Bell,
  ShieldCheck,
  Clock,
  AlertCircle,
  Check,
  CreditCard,
  Banknote,
  Ruler,
  ArrowUpRight,
  X,
} from "lucide-react";
import { Stars } from "@/shared/ui/stars";
import { CustomButton } from "@/shared/ui";
import QuantitySelector from "@/shared/ui/quantitySelector";
import { usePublicSettings } from "@/modules/settings/model/publicSettingsContext";
import { formatMoney } from "@/shared/lib/money";
import { resolveDeliveryWindow } from "../../utils/productHelpers";
import { ProductVariationAxis, ProductVariant } from "../../types";

// Named color swatches (EN + AR) with hex fallbacks; unknown values get a neutral swatch.
const COLOR_SWATCHES: Record<string, string> = {
  black: "#111111",
  white: "#ffffff",
  blue: "#2563eb",
  navy: "#1e3a8a",
  "navy blue": "#1e3a8a",
  red: "#dc2626",
  green: "#16a34a",
  gray: "#9ca3af",
  grey: "#9ca3af",
  beige: "#e8d8c3",
  yellow: "#facc15",
  orange: "#f97316",
  purple: "#7c3aed",
  pink: "#ec4899",
  brown: "#92400e",
  gold: "#d4af37",
  silver: "#c0c0c0",
  maroon: "#7f1d1d",
  teal: "#0d9488",
  olive: "#6b7a1f",
  cream: "#f5efe0",
  "light blue": "#93c5fd",
  "dark blue": "#1e40af",
  أسود: "#111111",
  أبيض: "#ffffff",
  أزرق: "#2563eb",
  أحمر: "#dc2626",
  أخضر: "#16a34a",
  رمادي: "#9ca3af",
  بيج: "#e8d8c3",
  أصفر: "#facc15",
  برتقالي: "#f97316",
  بنفسجي: "#7c3aed",
  بني: "#92400e",
  وردي: "#ec4899",
};

function resolveSwatch(value: string): string {
  const normalized = value.trim().toLowerCase();
  if (/^#([0-9a-f]{3,4}|[0-9a-f]{6}|[0-9a-f]{8})$/i.test(normalized)) {
    return normalized;
  }
  return COLOR_SWATCHES[normalized] ?? "#d4d4d8";
}

function isLightSwatch(hex: string): boolean {
  let clean = hex.replace("#", "");
  if (clean.length === 3) clean = clean.split("").map((c) => c + c).join("");
  const r = parseInt(clean.slice(0, 2), 16);
  const g = parseInt(clean.slice(2, 4), 16);
  const b = parseInt(clean.slice(4, 6), 16);
  return (r * 299 + g * 587 + b * 114) / 1000 > 160;
}

// Standard unisex clothing measurements in centimeters.
const SIZE_CHART = [
  { size: "XS", chest: 84, waist: 66, hips: 90, length: 66 },
  { size: "S", chest: 90, waist: 72, hips: 96, length: 68 },
  { size: "M", chest: 96, waist: 78, hips: 102, length: 70 },
  { size: "L", chest: 102, waist: 84, hips: 108, length: 72 },
  { size: "XL", chest: 110, waist: 92, hips: 116, length: 74 },
  { size: "XXL", chest: 118, waist: 100, hips: 124, length: 76 },
];

interface ProductActionsProps {
  productCategory: string;
  productName: string;
  stars?: number;
  reviewsCount?: number;
  price: number;
  oldPrice?: number;
  currency: string;
  variationAxes?: ProductVariationAxis[];
  variants?: ProductVariant[];
  selectedValues: Record<string, string>;
  onSelectAxis: (key: string, value: string) => void;
  availableStock: number;
  isVariantInStock: boolean;
  quantity: number;
  onQuantityChange: (val: number) => void;
  /** Per-unit volume-discount saving (0 when no tier applies). */
  tierSavingsPerUnit?: number;
  tierMinQuantity?: number;
  actionError?: string | null;
  isAdding: boolean;
  /** Add-to-cart just succeeded; button shows a temporary success state. */
  justAdded?: boolean;
  isAuthenticated: boolean;
  isWishlisted: boolean;
  onAddToCart: () => void;
  onBuyNow: () => void;
  onToggleWishlist: () => void;
  onShare: () => void;
  /** Opens the notify-me dialog. Rendered only when the variant is out of stock. */
  onNotifyMe?: () => void;
}

export function ProductActions({
  productCategory,
  productName,
  stars = 0,
  reviewsCount = 0,
  price,
  oldPrice,
  currency,
  variationAxes = [],
  variants = [],
  selectedValues,
  onSelectAxis,
  availableStock,
  isVariantInStock,
  quantity,
  onQuantityChange,
  tierSavingsPerUnit = 0,
  tierMinQuantity,
  actionError,
  isAdding,
  justAdded,
  isAuthenticated,
  isWishlisted,
  onAddToCart,
  onBuyNow,
  onToggleWishlist,
  onShare,
  onNotifyMe,
}: ProductActionsProps) {
  const t = useTranslations("products");
  const locale = useLocale();
  const { settings } = usePublicSettings();
  const roundedStars = Number(stars) || 0;

  const sizeDialogRef = useRef<HTMLDialogElement>(null);
  const [sizeGuideOpen, setSizeGuideOpen] = useState(false);

  useEffect(() => {
    const dialog = sizeDialogRef.current;
    if (!dialog) return;
    if (sizeGuideOpen) {
      if (!dialog.open) dialog.showModal();
      const previousOverflow = document.body.style.overflow;
      document.body.style.overflow = "hidden";
      return () => {
        document.body.style.overflow = previousOverflow;
      };
    }
    if (dialog.open) dialog.close();
  }, [sizeGuideOpen]);

  const isVariantPurchasable = (variant: ProductVariant) => {
    const stockCount = variant.stock?.availableQuantity ?? variant.availableStock;
    if (variant.stock?.isInStock === false || variant.inStock === false) return false;
    if (typeof stockCount === "number") return stockCount > 0;
    return variant.stock?.isInStock ?? variant.inStock ?? true;
  };

  // An option is purchasable when at least one matching variant (with the
  // other selected axes) is in stock.
  const isOptionAvailable = (axisKey: string, value: string) => {
    if (variants.length === 0) return true;
    const nextSelection = { ...selectedValues, [axisKey]: value };
    return variants.some((variant) => {
      const attributes = variant.attributes ?? {};
      const axisValue = attributes[axisKey];
      if (axisValue === undefined || axisValue === null || String(axisValue) !== value) return false;
      const matchesOtherAxes = variationAxes.every((axis) => {
        if (axis.key === axisKey) return true;
        const selected = nextSelection[axis.key];
        if (!selected) return true;
        const otherValue = attributes[axis.key];
        return otherValue !== undefined && otherValue !== null && String(otherValue) === selected;
      });
      return matchesOtherAxes && isVariantPurchasable(variant);
    });
  };

  const deliveryWindow = resolveDeliveryWindow(settings?.estimatedDeliveryDays, locale);
  const stockState =
    !isVariantInStock || availableStock <= 0 ? "out" : availableStock <= 5 ? "low" : "in";

  return (
    <div className="flex flex-col">
      <div className="mb-4 flex flex-col gap-2">
        <span className="w-fit text-[11px] font-semibold uppercase tracking-[0.14em] text-accent-brand">
          {productCategory.replace("-", " ")}
        </span>
        <h1 className="text-start text-2xl font-bold leading-tight text-foreground sm:text-3xl lg:text-4xl">
          {productName}
        </h1>
      </div>

      <div className="mb-5 flex flex-wrap items-center gap-3">
        <button
          type="button"
          onClick={() => document.getElementById("reviews")?.scrollIntoView({ behavior: "smooth" })}
          className="flex items-center gap-1.5 cursor-pointer"
          aria-label={t("actions.reviewsCount", { count: reviewsCount })}
        >
          <Stars starsNum={roundedStars} />
          {roundedStars > 0 && (
            <span className="ms-1 text-sm font-semibold text-foreground">
              {roundedStars.toFixed(1)}
            </span>
          )}
          <span className="ms-1 text-sm text-muted-foreground underline decoration-dashed underline-offset-4 transition-colors hover:text-accent-brand">
            {t("actions.reviewsCount", { count: reviewsCount })}
          </span>
        </button>
      </div>

      <div className="mb-6 flex flex-wrap items-end gap-x-3 gap-y-1 border-y border-border/70 py-4">
        <span className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
          {formatMoney(price, currency)}
        </span>
        {oldPrice !== undefined && oldPrice > 0 && oldPrice !== price && (
          <span className="mb-1 text-lg text-muted-foreground line-through">
            {formatMoney(oldPrice, currency)}
          </span>
        )}
        <span
          aria-live="polite"
          className={`mb-1 ms-auto inline-flex items-center gap-2 text-xs font-semibold ${
            stockState === "out"
              ? "text-destructive"
              : stockState === "low"
                ? "rounded-full bg-amber-500/10 px-2.5 py-1 text-amber-700 dark:text-amber-400"
                : "text-emerald-700 dark:text-emerald-400"
          }`}
        >
          <span
            aria-hidden="true"
            className={`size-1.5 rounded-full ${
              stockState === "out"
                ? "bg-destructive"
                : stockState === "low"
                  ? "bg-amber-500"
                  : "bg-emerald-600"
            }`}
          />
          {stockState === "out"
            ? t("actions.outOfStock")
            : stockState === "low"
              ? t("actions.lowStock", { count: availableStock })
              : t("actions.inStock")}
        </span>
      </div>

      {tierSavingsPerUnit > 0 && (
        <p className="mb-4 inline-flex items-center gap-2 rounded-full bg-emerald-500/10 px-3.5 py-1.5 text-xs font-semibold text-emerald-600 dark:text-emerald-400">
          {t("tierPricing.activeSavings", {
            amount: formatMoney(tierSavingsPerUnit, currency),
            minQuantity: tierMinQuantity ?? quantity,
          })}
        </p>
      )}

      {/* Variation axes (one selector per theme attribute) */}
      {variationAxes.map((axis) => {
        const selectedValue = selectedValues[axis.key] || "";
        const label = axis.key === "color" ? t("actions.color") : axis.label;
        const isColorAxis = axis.key === "color";
        const isSizeAxis = axis.key.toLowerCase() === "size";
        return (
          <fieldset key={axis.key} className="mb-5 flex flex-col gap-2.5">
            <legend className="text-sm font-semibold text-foreground">
              {label}{" "}
              <span className="ms-1 font-normal text-muted-foreground">
                {selectedValue}
              </span>
              {isSizeAxis && (
                <button
                  type="button"
                  onClick={() => setSizeGuideOpen(true)}
                  className="ms-2 inline-flex cursor-pointer items-center gap-1 text-xs font-medium text-accent-brand transition-colors hover:underline"
                >
                  <Ruler className="size-3.5" aria-hidden="true" />
                  {t("sizeGuide.open")}
                  <ArrowUpRight className="size-3 rtl:-scale-x-100" aria-hidden="true" />
                </button>
              )}
            </legend>
            <div className="flex flex-wrap gap-2" role="radiogroup" aria-label={label}>
              {axis.values.map((value) => {
                const isSelected = selectedValue === value;
                const isAvailable = isOptionAvailable(axis.key, value);
                const availabilityTitle = isAvailable
                  ? value
                  : `${value} — ${t("actions.outOfStock")}`;

                if (isColorAxis) {
                  const swatch = resolveSwatch(value);
                  const light = isLightSwatch(swatch);
                  return (
                    <button
                      key={value}
                      type="button"
                      role="radio"
                      aria-checked={isSelected}
                      aria-disabled={!isAvailable}
                      disabled={!isAvailable}
                      aria-label={value}
                      title={availabilityTitle}
                      onClick={() => onSelectAxis(axis.key, value)}
                      style={{ backgroundColor: swatch }}
                      className={`flex size-9 items-center justify-center rounded-full border border-border/70 transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent-brand focus-visible:ring-offset-2 ${
                        !isAvailable
                          ? "cursor-not-allowed opacity-40"
                          : isSelected
                            ? "cursor-pointer ring-2 ring-foreground ring-offset-2 ring-offset-background"
                            : "cursor-pointer hover:scale-105 hover:ring-1 hover:ring-foreground/40 hover:ring-offset-2"
                      }`}
                    >
                      {isSelected && (
                        <Check
                          className={`size-4 ${light ? "text-foreground" : "text-white"}`}
                          aria-hidden="true"
                        />
                      )}
                    </button>
                  );
                }

                return (
                  <button
                    key={value}
                    type="button"
                    role="radio"
                    aria-checked={isSelected}
                    aria-disabled={!isAvailable}
                    disabled={!isAvailable}
                    title={!isAvailable ? t("actions.outOfStock") : undefined}
                    onClick={() => onSelectAxis(axis.key, value)}
                    className={`min-h-11 min-w-11 rounded-lg border px-3.5 py-2 text-sm transition-colors ${
                      !isAvailable
                        ? "cursor-not-allowed border-border bg-muted/30 text-muted-foreground line-through opacity-50"
                        : isSelected
                          ? "cursor-pointer border-foreground bg-foreground font-semibold text-background"
                          : "cursor-pointer border-border bg-background text-foreground hover:border-foreground/50"
                    }`}
                  >
                    {value}
                  </button>
                );
              })}
            </div>
          </fieldset>
        );
      })}

      {/* Purchase actions */}
      <div className="flex flex-col gap-3">
        <div className="h-12 w-full">
          <QuantitySelector
            value={Math.min(quantity, Math.max(1, availableStock))}
            max={Math.max(1, availableStock)}
            disabled={!isVariantInStock}
            onChange={onQuantityChange}
          />
        </div>

        {quantity > 1 && (
          <p className="text-xs font-medium text-muted-foreground" aria-live="polite">
            {t("actions.lineTotal", {
              total: formatMoney(price * quantity, currency),
            })}
          </p>
        )}

        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <CustomButton
            text={
              isAdding
                ? t("actions.adding")
                : justAdded
                  ? t("actions.added")
                  : isVariantInStock
                    ? t("actions.addToCart")
                    : t("actions.outOfStock")
            }
            onClick={onAddToCart}
            className={`py-3.5 text-sm ${
              !isVariantInStock
                ? "cursor-not-allowed opacity-50"
                : justAdded
                  ? "bg-green-600 text-white hover:bg-green-600"
                  : ""
            }`}
            disabled={isAdding || !isVariantInStock}
          />

          <button
            type="button"
            onClick={onBuyNow}
            disabled={isAdding || !isVariantInStock}
            className="cursor-pointer rounded-lg bg-accent-brand px-4 py-3.5 text-sm font-semibold text-white transition-colors hover:bg-accent-brand/90 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {t("actions.buyNow")}
          </button>
        </div>

        {actionError && (
          <div
            role="alert"
            className="rounded-xl border border-destructive/30 bg-destructive/10 px-3.5 py-2.5 text-xs font-semibold text-destructive flex items-center gap-2"
          >
            <AlertCircle className="size-4 shrink-0" aria-hidden="true" />
            <span>{actionError}</span>
          </div>
        )}

        <div className="flex flex-wrap items-center gap-2 text-[11px] text-muted-foreground">
          <span className="inline-flex items-center gap-1.5 rounded-full border border-border/70 bg-card px-2.5 py-1">
            <ShieldCheck className="size-3.5" aria-hidden="true" />
            {t("trustPill.secureCheckout")}
          </span>
          {settings?.allowCards && (
            <span className="inline-flex items-center gap-1.5 rounded-full border border-border/70 bg-card px-2.5 py-1">
              <CreditCard className="size-3.5" aria-hidden="true" />
              {t("trustPill.cards")}
            </span>
          )}
          {settings?.allowCod && (
            <span className="inline-flex items-center gap-1.5 rounded-full border border-border/70 bg-card px-2.5 py-1">
              <Banknote className="size-3.5" aria-hidden="true" />
              {t("trustPill.cashOnDelivery")}
            </span>
          )}
          {settings?.estimatedDeliveryDays?.trim() && (
            <span
              suppressHydrationWarning
              className="inline-flex items-center gap-1.5 rounded-full border border-border/70 bg-card px-2.5 py-1"
            >
              <Clock className="size-3.5" aria-hidden="true" />
              {deliveryWindow
                ? deliveryWindow.start === deliveryWindow.end
                  ? t("trustPill.arrivesOn", { date: deliveryWindow.start })
                  : t("trustPill.arrives", {
                      start: deliveryWindow.start,
                      end: deliveryWindow.end,
                    })
                : t("trustPill.deliveryDays", { days: settings.estimatedDeliveryDays })}
            </span>
          )}
        </div>

        {!isVariantInStock && onNotifyMe && (
          <button
            type="button"
            onClick={onNotifyMe}
            className="flex min-h-11 w-full items-center justify-center gap-2 rounded-lg bg-accent-brand px-4 py-2.5 text-xs font-semibold text-white transition-opacity hover:opacity-90"
          >
            <Bell className="size-3.5" aria-hidden="true" />
            {t("notifyMe.open")}
          </button>
        )}

        <div className="flex gap-3">
          {isAuthenticated && (
            <button
              type="button"
              onClick={onToggleWishlist}
              className={`min-h-11 flex-1 cursor-pointer rounded-lg border px-4 py-2.5 text-xs font-medium transition-colors ${
                isWishlisted
                  ? "border-foreground/30 bg-secondary text-foreground"
                  : "border-border bg-background text-muted-foreground hover:border-foreground/30 hover:text-foreground"
              }`}
            >
              {isWishlisted ? t("actions.savedToWishlist") : t("actions.addToWishlist")}
            </button>
          )}

          <button
            type="button"
            onClick={onShare}
            className="min-h-11 flex-1 cursor-pointer rounded-lg border border-border bg-background px-4 py-2.5 text-xs font-medium text-muted-foreground transition-colors hover:border-foreground/30 hover:text-foreground"
          >
            {t("actions.share")}
          </button>
        </div>
      </div>

      <div className="mt-6 grid grid-cols-1 gap-4 border-t border-border/70 pt-5 sm:grid-cols-2">
        <div>
          <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-muted-foreground">
            {t("actions.freeDelivery")}
          </p>
          <p className="mt-1 text-sm text-foreground">
            {t("actions.freeDeliveryHint", {
              threshold: settings ? formatMoney(settings.freeShippingThreshold, settings.currency?.trim() ?? "") : "—",
            })}
          </p>
        </div>
        <div>
          <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-muted-foreground">
            {t("actions.easyReturns")}
          </p>
          <p className="mt-1 text-sm text-foreground">{t("actions.easyReturnsHint")}</p>
        </div>
      </div>

      <dialog
        ref={sizeDialogRef}
        aria-label={t("sizeGuide.title")}
        aria-modal="true"
        onCancel={(event) => {
          event.preventDefault();
          setSizeGuideOpen(false);
        }}
        onClose={() => setSizeGuideOpen(false)}
        onClick={(event) => {
          if (event.target === event.currentTarget) setSizeGuideOpen(false);
        }}
        className="m-auto w-[calc(100%-1rem)] max-w-xl overflow-hidden rounded-2xl border border-border bg-card p-0 text-foreground shadow-2xl backdrop:bg-foreground/70 backdrop:backdrop-blur-sm"
      >
        <div className="flex items-center justify-between gap-3 border-b border-border px-4 py-3">
          <h2 className="text-base font-bold text-foreground">{t("sizeGuide.title")}</h2>
          <button
            type="button"
            onClick={() => setSizeGuideOpen(false)}
            aria-label={t("sizeGuide.close")}
            className="flex size-11 cursor-pointer items-center justify-center rounded-full text-foreground transition-colors hover:bg-secondary"
          >
            <X className="size-5" aria-hidden="true" />
          </button>
        </div>

        <div className="max-h-[70vh] overflow-auto p-4">
          <table className="w-full min-w-[320px] border-collapse text-sm">
            <thead>
              <tr className="border-b border-border text-start text-xs uppercase tracking-wide text-muted-foreground">
                <th scope="col" className="px-2 py-2 text-start">{t("sizeGuide.size")}</th>
                <th scope="col" className="px-2 py-2 text-start">{t("sizeGuide.chest")}</th>
                <th scope="col" className="px-2 py-2 text-start">{t("sizeGuide.waist")}</th>
                <th scope="col" className="px-2 py-2 text-start">{t("sizeGuide.hips")}</th>
                <th scope="col" className="px-2 py-2 text-start">{t("sizeGuide.length")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/70">
              {SIZE_CHART.map((row) => (
                <tr key={row.size}>
                  <th scope="row" className="px-2 py-2 text-start font-semibold text-foreground">
                    {row.size}
                  </th>
                  <td className="px-2 py-2 text-foreground/90">{row.chest}</td>
                  <td className="px-2 py-2 text-foreground/90">{row.waist}</td>
                  <td className="px-2 py-2 text-foreground/90">{row.hips}</td>
                  <td className="px-2 py-2 text-foreground/90">{row.length}</td>
                </tr>
              ))}
            </tbody>
          </table>
          <p className="mt-3 text-xs leading-relaxed text-muted-foreground">
            {t("sizeGuide.note")}
          </p>
        </div>
      </dialog>
    </div>
  );
}
