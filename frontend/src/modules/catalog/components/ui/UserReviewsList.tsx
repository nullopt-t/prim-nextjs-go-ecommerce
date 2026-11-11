"use client";

import { useCallback, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { catalogService } from "@/modules/catalog/api/catalogService";
import { mapReview } from "@/modules/catalog/utils/mapReview";
import { Stars } from "@/shared/ui/stars";
import { Link } from "@/shared/i18n/navigation";
import { REVIEWS_PAGE_SIZE } from "@/shared/lib/reviewQuery";
import type { ProductReview, Review } from "../../types";

interface UserReviewsListProps {
  /** Author whose public review profile this page shows. */
  userId: string;
  /** First page, mapped on the server so the profile renders in the HTML. */
  initialReviews: ProductReview[];
  initialHasMore: boolean;
  initialCursor: string | null;
}

/**
 * The reviewer's public profile list: every review they wrote, each linked to
 * the product it was written about. Mirrors the product detail page's pager:
 * the first page arrives from the server, "show more" appends cursor pages.
 */
export function UserReviewsList({
  userId,
  initialReviews,
  initialHasMore,
  initialCursor,
}: UserReviewsListProps) {
  const t = useTranslations("reviewerProfile");
  const locale = useLocale();
  const [reviews, setReviews] = useState(initialReviews);
  const [cursor, setCursor] = useState(initialCursor);
  const [hasMore, setHasMore] = useState(initialHasMore);
  const [loadingMore, setLoadingMore] = useState(false);

  const handleLoadMore = useCallback(async () => {
    if (loadingMore) return;

    setLoadingMore(true);
    try {
      const res = await catalogService.getUserReviews(userId, {
        cursor,
        pageSize: REVIEWS_PAGE_SIZE,
      });
      const list = (res?.data ?? []).map((r: Review) =>
        mapReview(r, {
          locale,
          authorLabel: t("anonymous"),
          titleLabel: t("noTitle"),
          dateLabel: t("recently"),
        })
      );
      const meta = res?.meta as { nextCursor?: string; hasMore?: boolean } | undefined;
      setReviews((prev) => {
        const seen = new Set(prev.map((review) => String(review.id)));
        const additions = list.filter((review) => !seen.has(String(review.id)));
        return additions.length ? [...prev, ...additions] : prev;
      });
      setCursor(meta?.nextCursor ?? null);
      setHasMore(Boolean(meta?.hasMore));
    } catch (err) {
      console.error("Failed to load more reviews:", err);
    } finally {
      setLoadingMore(false);
    }
  }, [userId, cursor, loadingMore, locale, t]);

  if (reviews.length === 0) {
    return (
      <div className="mt-6 flex flex-col items-center justify-center rounded-xl border border-dashed border-border/70 bg-secondary/25 px-6 py-12 text-center">
        <p className="text-sm font-semibold text-foreground">{t("empty")}</p>
      </div>
    );
  }

  return (
    <div className="mt-6 flex flex-col gap-3">
      {reviews.map((review) => (
        <article
          key={review.id}
          className="rounded-xl border border-border/70 bg-secondary/25 p-5"
        >
          {review.productId && review.productTitle ? (
            <Link
              href={`/products/${review.productId}`}
              dir="auto"
              className="text-start text-sm font-semibold text-accent-brand underline-offset-4 transition-colors hover:underline"
            >
              {review.productTitle}
            </Link>
          ) : null}

          <div className="mt-2 flex flex-wrap items-center gap-2">
            <Stars starsNum={review.rating} />
            <span className="text-xs font-medium text-muted-foreground">
              {Number(review.rating).toFixed(1)}
            </span>
            <time className="ms-auto text-[11px] text-muted-foreground">{review.date}</time>
          </div>

          <h3 dir="auto" className="mt-3 text-start text-sm font-bold text-foreground">
            {review.title}
          </h3>
          <p
            dir="auto"
            className="mt-1.5 whitespace-pre-line text-start text-sm leading-relaxed text-foreground/85"
          >
            {review.content}
          </p>
        </article>
      ))}

      {hasMore && (
        <button
          type="button"
          onClick={() => handleLoadMore()}
          disabled={loadingMore}
          className="mx-auto mt-1 cursor-pointer rounded-xl border border-border bg-background px-6 py-2.5 text-sm font-medium text-foreground shadow-sm transition-colors hover:bg-secondary disabled:cursor-not-allowed disabled:opacity-60"
        >
          {loadingMore ? t("loadingMore") : t("showMore")}
        </button>
      )}
    </div>
  );
}
