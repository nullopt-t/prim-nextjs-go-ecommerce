"use client";

import { useEffect, useState } from "react";
import { Link } from "@/i18n/navigation";
import { userService } from "@/services/user";
import { Stars } from "@/components/ui/stars";
import { useTranslations, useLocale } from "next-intl";
import { notify, apiErrorMessage } from "@/lib/notify";

// ── Types ──────────────────────────────────────────────────────────────────────
interface Review {
  id: string;
  productId: string;
  userId: string;
  rating: number;
  title?: string;
  body?: string;
  status: "pending" | "approved" | "rejected";
  createdAt: string;
  updatedAt: string;
}

// ── Status badge map ───────────────────────────────────────────────────────────
const STATUS_BADGE: Record<Review["status"], { key: string; className: string }> = {
  pending: {
    key: "pending",
    className:
      "text-amber-600 bg-amber-50 dark:bg-amber-950/30 dark:text-amber-400",
  },
  approved: {
    key: "approved",
    className:
      "text-emerald-600 bg-emerald-50 dark:bg-emerald-950/30 dark:text-emerald-400",
  },
  rejected: {
    key: "rejected",
    className: "text-red-600 bg-red-50 dark:bg-red-950/30 dark:text-red-400",
  },
};

// ── Skeleton card ──────────────────────────────────────────────────────────────
function ReviewSkeleton() {
  return <div className="animate-pulse bg-secondary/40 rounded-xl h-28 w-full border border-border/40" />;
}

// ── Single review card ─────────────────────────────────────────────────────────
function ReviewCard({
  review,
  onDelete,
}: {
  review: Review;
  onDelete: (id: string) => void;
}) {
  const [expanded, setExpanded] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const t = useTranslations("userReviews");
  const locale = useLocale();
  const badgeCfg = STATUS_BADGE[review.status] ?? STATUS_BADGE.pending;
  const date = new Date(review.createdAt).toLocaleDateString(locale);
  const bodyTruncated = review.body && review.body.length > 160 && !expanded;

  const handleDelete = async () => {
    setDeleting(true);
    try {
      await userService.deleteReview(review.id);
      notify.success(t("toastDeleted"));
      onDelete(review.id);
    } catch (err: unknown) {
      notify.error(apiErrorMessage(err, t("toastFailed")));
    } finally {
      setDeleting(false);
    }
  };

  return (
    <div className="bg-secondary/30 hover:bg-secondary/40 border border-border/50 rounded-xl shadow-2xs p-3.5 sm:p-4 flex flex-col gap-2.5 transition-colors">
      {/* Top row */}
      <div className="flex items-start justify-between gap-3">
        <div className="flex flex-col gap-1 min-w-0">
          <div className="flex items-center gap-2">
            <Stars starsNum={review.rating} />
            <span className="text-[11px] text-muted-foreground">• {date}</span>
          </div>
          {review.title && (
            <p className="text-xs sm:text-sm font-semibold text-foreground truncate">{review.title}</p>
          )}
        </div>
        <div className="flex items-center gap-2 shrink-0">
          <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold ${badgeCfg.className}`}>
            {t(`status.${badgeCfg.key}`)}
          </span>
          <button
            type="button"
            onClick={handleDelete}
            disabled={deleting}
            aria-label={t("ariaDelete")}
            className="text-xs font-medium text-destructive hover:underline transition disabled:opacity-40 cursor-pointer"
          >
            {deleting ? t("deleting") : t("delete")}
          </button>
        </div>
      </div>

      {/* Body */}
      {review.body && (
        <div>
          <p
            className={`text-xs text-muted-foreground leading-relaxed ${bodyTruncated ? "line-clamp-3" : ""}`}
          >
            {review.body}
          </p>
          {review.body.length > 160 && (
            <button
              type="button"
              onClick={() => setExpanded((v) => !v)}
              className="text-xs font-medium text-accent-brand hover:underline mt-1 cursor-pointer"
            >
              {expanded ? t("showLess") : t("showMore")}
            </button>
          )}
        </div>
      )}
    </div>
  );
}

// ── Main component ─────────────────────────────────────────────────────────────
export default function ReviewsContent() {
  const [reviews, setReviews] = useState<Review[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const t = useTranslations("userReviews");

  useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);
      try {
        const res = await userService.getMyReviews();
        if (cancelled) return;
        const raw: Review[] = Array.isArray(res) ? res : res?.data ?? [];
        setReviews(raw);
      } catch (err: unknown) {
        if (!cancelled) setError(err instanceof Error ? err.message : "Failed to load reviews.");
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    load();
    return () => {
      cancelled = true;
    };
  }, []);

  const handleDelete = (id: string) => {
    setReviews((prev) => prev.filter((r) => r.id !== id));
  };

  return (
    <div className="flex flex-col gap-6">
      {/* Page title */}
      <div>
        <h1 className="text-xl font-semibold text-foreground">{t("title")}</h1>
        <p className="text-sm text-muted-foreground mt-1">
          {t("subtitle")}
        </p>
      </div>

      {/* Content */}
      {loading ? (
        <div className="flex flex-col gap-3">
          <ReviewSkeleton />
          <ReviewSkeleton />
        </div>
      ) : error ? (
        <p className="text-destructive text-sm">Error: {error}</p>
      ) : reviews.length === 0 ? (
        <div className="bg-card border border-border rounded-lg p-12 flex flex-col items-center justify-center text-center gap-2 shadow-sm">
          <p className="font-medium text-foreground">{t("empty.title")}</p>
          <p className="text-sm text-muted-foreground max-w-sm">
            {t("empty.subtitle")}
          </p>
          <Link
            href="/products"
            className="mt-3 h-8 px-4 inline-flex items-center rounded-md border border-border text-sm font-medium hover:bg-secondary transition"
          >
            {t("empty.action")}
          </Link>
        </div>
      ) : (
        <div className="flex flex-col gap-3">
          {reviews.map((review) => (
            <ReviewCard key={review.id} review={review} onDelete={handleDelete} />
          ))}
        </div>
      )}
    </div>
  );
}
