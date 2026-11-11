import { useState } from "react";
import { Star, AlertCircle } from "lucide-react";
import { Stars } from "@/shared/ui/stars";
import { useTranslations } from "next-intl";
import { ProductReview } from "../../types";
import { CatalogImage } from "@/shared/ui/catalogImage";
import { useAuthContext } from "@/modules/auth/model/authContext";

interface ProductReviewsProps {
  stars?: number;
  reviewsCount?: number;
  reviews: ProductReview[];
  canReview?: boolean;
  onAddReview: (review: { rating: number; title: string; content: string }) => Promise<{ success: boolean; error?: string } | boolean | void> | void;
  onVoteHelpful: (reviewId: string | number) => void;
  /**
   * Authoritative star breakdown from the backend, keyed by score ("1".."5").
   * The loaded review page is only a slice, so the histogram must not be
   * derived from it.
   */
  distribution?: Record<string, number>;
  /**
   * Active star filter, or null for every rating. The filter is applied by the
   * backend, so the list always matches the histogram counts.
   */
  ratingFilter?: number | null;
  onRatingFilterChange?: (rating: number | null) => void;
  /** Whether the backend holds more reviews for the active filter. */
  hasMore?: boolean;
  /** Fetches the next page of reviews when the shopper asks for more. */
  onLoadMore?: () => void;
  loadingMore?: boolean;
}

export function ProductReviews({
  stars = 0,
  reviewsCount = 0,
  reviews,
  canReview = false,
  onAddReview,
  onVoteHelpful,
  distribution,
  ratingFilter = null,
  onRatingFilterChange,
  hasMore = false,
  onLoadMore,
  loadingMore = false,
}: ProductReviewsProps) {
  const t = useTranslations("reviews");
  const { isAuthenticated } = useAuthContext();
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [form, setForm] = useState({ rating: 0, title: "", content: "" });
  const [formError, setFormError] = useState<string | null>(null);

  const handleSubmit = async () => {
    if (!form.rating || !form.title.trim() || !form.content.trim() || isSubmitting) return;
    setFormError(null);
    setIsSubmitting(true);
    try {
      const res = await onAddReview({
        rating: form.rating,
        title: form.title.trim(),
        content: form.content.trim(),
      });
      const ok = typeof res === "object" ? res.success : res !== false;
      const errMsg = typeof res === "object" && !res.success ? res.error : null;
      if (ok) {
        setForm({ rating: 0, title: "", content: "" });
        setIsFormOpen(false);
      } else if (errMsg) {
        setFormError(errMsg);
      }
    } catch {
      setFormError(t("submitFailed"));
    } finally {
      setIsSubmitting(false);
    }
  };

  const closeForm = () => {
    setIsFormOpen(false);
    setForm({ rating: 0, title: "", content: "" });
    setFormError(null);
  };

  const loadedCount = reviews.length;
  const averageRating = Number(stars) || 0;
  const displayCount = reviewsCount || loadedCount;
  // The backend breakdown covers every approved review; fall back to the
  // loaded slice only when the backend did not send one.
  const backendTotal = distribution
    ? Object.values(distribution).reduce((sum, value) => sum + (Number(value) || 0), 0)
    : 0;
  const ratingCounts = [5, 4, 3, 2, 1].map((starValue) => ({
    rating: starValue,
    count: distribution
      ? Number(distribution[String(starValue)] ?? 0)
      : reviews.filter((review) => Math.round(Number(review.rating)) === starValue).length,
  }));
  const histogramTotal = backendTotal || loadedCount;

  const changeRatingFilter = (nextFilter: number | null) => {
    onRatingFilterChange?.(nextFilter);
  };

  return (
    <section className="scroll-mt-24 rounded-2xl border border-border bg-card p-5 shadow-sm sm:p-6">
      <div className="mb-5 flex flex-col gap-4 border-b border-border/70 pb-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <p className="mb-1 text-[11px] font-bold uppercase tracking-[0.16em] text-accent-brand">
            {t("sectionLabel")}
          </p>
          <h2 className="text-base font-semibold text-foreground sm:text-lg">
            {t("customerReviews")}
          </h2>
        </div>
        {isAuthenticated && canReview && !isFormOpen && (
          <button
            type="button"
            onClick={() => {
              setFormError(null);
              setIsFormOpen(true);
            }}
            className="w-fit cursor-pointer rounded-xl bg-primary px-5 py-2.5 text-sm font-semibold text-primary-foreground transition-colors hover:bg-primary/90"
          >
            {t("writeReview")}
          </button>
        )}
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
        {/* Summary / Rating breakdown */}
        <aside className="lg:col-span-4">
          <div className="lg:sticky lg:top-24">
            <div className="flex items-end gap-3">
              <span className="text-5xl font-bold leading-none text-foreground">
                {averageRating.toFixed(1)}
              </span>
              <div className="pb-1">
                <Stars starsNum={averageRating} />
                {averageRating > 0 && (
                  <p className="mt-1 text-xs text-muted-foreground">
                    {t("outOfFive", { stars: averageRating.toFixed(1) })}
                  </p>
                )}
              </div>
            </div>

            <p className="mt-3 text-sm text-muted-foreground">
              {t("globalRatings", { count: displayCount })}
            </p>

            <div className="mt-5 flex flex-col gap-2">
              {ratingCounts.map(({ rating, count }) => {
                const pct = histogramTotal > 0 ? Math.round((count / histogramTotal) * 100) : 0;
                const isActive = ratingFilter === rating;
                return (
                  <button
                    key={rating}
                    type="button"
                    onClick={() => changeRatingFilter(isActive ? null : rating)}
                    aria-label={t("starLabel", { star: rating })}
                    aria-pressed={isActive}
                    className={`group flex items-center gap-2.5 rounded-lg px-1 py-0.5 text-xs transition-colors cursor-pointer ${
                      isActive ? "text-foreground" : "text-muted-foreground hover:text-foreground"
                    }`}
                  >
                    <span className="w-7 shrink-0 text-start font-medium">{rating} ★</span>
                    <span className="h-2.5 flex-1 overflow-hidden rounded-full bg-secondary">
                      <span
                        className={`block h-full rounded-full transition-all duration-500 ${
                          isActive ? "bg-accent-brand" : "bg-amber-400"
                        }`}
                        style={{ width: `${pct}%` }}
                      />
                    </span>
                    <span className="w-9 text-end text-[11px] text-muted-foreground">{pct}%</span>
                  </button>
                );
              })}
            </div>

            {ratingFilter !== null && (
              <button
                type="button"
                onClick={() => changeRatingFilter(null)}
                className="mt-4 text-xs font-semibold text-accent-brand hover:underline cursor-pointer"
              >
                {t("clearFilters")}
              </button>
            )}

            {isAuthenticated && canReview && (
              <div className="mt-6 border-t border-border pt-5">
                <p className="text-sm font-semibold text-foreground">{t("reviewThisProduct")}</p>
                <p className="mt-1 text-xs leading-relaxed text-muted-foreground">
                  {t("shareThoughts")}
                </p>
                {!isFormOpen && (
                  <button
                    type="button"
                    onClick={() => {
                      setFormError(null);
                      setIsFormOpen(true);
                    }}
                    className="mt-3 w-full cursor-pointer rounded-xl border border-border bg-background px-4 py-2.5 text-sm font-medium text-foreground shadow-sm transition-colors hover:bg-secondary"
                  >
                    {t("writeShort")}
                  </button>
                )}
              </div>
            )}
          </div>
        </aside>

        {/* Review form and list */}
        <div className="flex flex-col gap-4 lg:col-span-8">
          {isAuthenticated && canReview && isFormOpen && (
            <div className="rounded-xl border border-accent-brand/30 bg-secondary/25 p-5">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <h3 className="text-base font-bold text-foreground">{t("reviewFormTitle")}</h3>
                  <p className="mt-1 text-xs text-muted-foreground">{t("reviewFormHint")}</p>
                </div>
                <button
                  type="button"
                  onClick={closeForm}
                  className="shrink-0 cursor-pointer text-xs font-medium text-muted-foreground transition-colors hover:text-foreground"
                >
                  {t("close")}
                </button>
              </div>

              <div className="mt-5 flex flex-col gap-4">
                <div>
                  <label className="mb-2 block text-sm font-semibold text-foreground">
                    {t("rating")}
                  </label>
                  <div className="flex items-center gap-1">
                    {[1, 2, 3, 4, 5].map((star) => (
                      <button
                        key={star}
                        type="button"
                        onClick={() => setForm({ ...form, rating: star })}
                        aria-label={t("starLabel", { star })}
                        aria-pressed={form.rating >= star}
                        className={`cursor-pointer p-0.5 ${
                          form.rating >= star ? "text-amber-500" : "text-border"
                        }`}
                      >
                        <Star
                          className="size-7"
                          fill={form.rating >= star ? "currentColor" : "none"}
                          stroke="currentColor"
                        />
                      </button>
                    ))}
                  </div>
                </div>

                <div>
                  <label htmlFor="review-headline" className="mb-1.5 block text-sm font-semibold text-foreground">
                    {t("addHeadline")}
                  </label>
                  <input
                    id="review-headline"
                    type="text"
                    dir="auto"
                    placeholder={t("headlinePlaceholder")}
                    className="w-full rounded-xl border border-border bg-background px-3.5 py-2.5 text-sm text-foreground transition-all focus:border-transparent focus:outline-none focus:ring-2 focus:ring-accent-brand"
                    value={form.title}
                    onChange={(e) => setForm({ ...form, title: e.target.value })}
                  />
                </div>

                <div>
                  <label htmlFor="review-body" className="mb-1.5 block text-sm font-semibold text-foreground">
                    {t("addWrittenReview")}
                  </label>
                  <textarea
                    id="review-body"
                    dir="auto"
                    placeholder={t("contentPlaceholder")}
                    className="min-h-[120px] w-full resize-y rounded-xl border border-border bg-background px-3.5 py-2.5 text-sm text-foreground transition-all focus:border-transparent focus:outline-none focus:ring-2 focus:ring-accent-brand"
                    value={form.content}
                    onChange={(e) => setForm({ ...form, content: e.target.value })}
                  />
                </div>

                {formError && (
                  <div role="alert" className="flex items-start gap-2 rounded-xl border border-destructive/30 bg-destructive/10 px-3.5 py-2.5 text-sm text-destructive">
                    <AlertCircle className="mt-0.5 size-4 shrink-0" aria-hidden="true" />
                    <span>{formError}</span>
                  </div>
                )}

                <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
                  <button
                    type="button"
                    onClick={closeForm}
                    className="cursor-pointer rounded-xl border border-border bg-background px-4 py-2.5 text-sm font-medium text-foreground shadow-sm transition-colors hover:bg-secondary"
                  >
                    {t("cancel")}
                  </button>
                  <button
                    type="button"
                    onClick={handleSubmit}
                    disabled={!form.rating || !form.title.trim() || !form.content.trim() || isSubmitting}
                    className="cursor-pointer rounded-xl bg-primary px-5 py-2.5 text-sm font-semibold text-primary-foreground shadow-sm transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    {isSubmitting ? t("submitting") : t("submit")}
                  </button>
                </div>
              </div>
            </div>
          )}

          {reviews.length === 0 ? (
            <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border/70 bg-secondary/25 px-6 py-12 text-center">
              <p className="text-base font-bold text-foreground">{t("noReviewsYet")}</p>
              <p className="mt-1 max-w-sm text-sm text-muted-foreground">{t("beTheFirst")}</p>
              {isAuthenticated && canReview && !isFormOpen && (
                <button
                  type="button"
                  onClick={() => {
                    setFormError(null);
                    setIsFormOpen(true);
                  }}
                  className="mt-5 cursor-pointer rounded-xl bg-primary px-5 py-2.5 text-sm font-semibold text-primary-foreground shadow-sm transition-colors hover:bg-primary/90"
                >
                  {t("writeReview")}
                </button>
              )}
            </div>
          ) : reviews.length === 0 ? (
            <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border/70 bg-secondary/25 px-6 py-10 text-center">
              <p className="text-sm font-semibold text-foreground">{t("noMatchingReviews")}</p>
              <button
                type="button"
                onClick={() => changeRatingFilter(null)}
                className="mt-3 cursor-pointer text-xs font-semibold text-accent-brand hover:underline"
              >
                {t("clearFilters")}
              </button>
            </div>
          ) : (
            <div className="flex flex-col gap-3">
              {reviews.map((review) => (
                <article
                  key={review.id}
                  className="rounded-xl border border-border/70 bg-secondary/25 p-5"
                >
                  <div className="flex items-start gap-3">
                    {review.avatar && (
                      <CatalogImage
                        src={review.avatar}
                        alt={review.author}
                        width={40}
                        height={40}
                        className="size-10 shrink-0 rounded-full object-cover"
                      />
                    )}

                    <div className="min-w-0 flex-1">
                      <div className="flex flex-wrap items-start justify-between gap-x-4 gap-y-1">
                        <div>
                          <p dir="auto" className="text-start text-sm font-semibold text-foreground">{review.author}</p>
                        </div>
                        <time className="text-[11px] text-muted-foreground">{review.date}</time>
                      </div>

                      <div className="mt-2 flex items-center gap-2">
                        <Stars starsNum={review.rating} />
                        <span className="text-xs font-medium text-muted-foreground">
                          {Number(review.rating).toFixed(1)}
                        </span>
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

                      <div className="mt-4 flex flex-wrap items-center gap-3 border-t border-border/70 pt-3">
                        {isAuthenticated && (
                          <button
                            type="button"
                            onClick={() => onVoteHelpful(review.id)}
                            aria-pressed={Boolean(review.userVoted)}
                            className={`cursor-pointer text-xs font-medium underline-offset-4 transition-colors hover:underline ${
                              review.userVoted
                                ? "text-accent-brand"
                                : "text-muted-foreground hover:text-foreground"
                            }`}
                          >
                            {t("helpful")}
                          </button>
                        )}
                        {Number(review.helpful) > 0 && (
                          <span className="text-[11px] text-muted-foreground">
                            {t("peopleFoundHelpful", { count: review.helpful })}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                </article>
              ))}
            </div>
          )}

          {hasMore && (
            <button
              type="button"
              onClick={() => onLoadMore?.()}
              disabled={loadingMore}
              className="mx-auto mt-1 cursor-pointer rounded-xl border border-border bg-background px-6 py-2.5 text-sm font-medium text-foreground shadow-sm transition-colors hover:bg-secondary disabled:cursor-not-allowed disabled:opacity-60"
            >
              {loadingMore ? t("loadingMore") : t("showMore")}
            </button>
          )}
        </div>
      </div>
    </section>
  );
}
