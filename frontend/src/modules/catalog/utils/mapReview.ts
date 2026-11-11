import type { ProductReview, Review } from "../types";

/**
 * Maps a backend review onto the PDP's view model. The server component and
 * the client component both render reviews, so the mapping lives here to keep
 * the two from drifting (the count and the list must agree).
 */
export function mapReview(review: Review, options: {
  locale: string;
  /** Label used when the review has no author name. */
  authorLabel: string;
  /** Label used for the signed-in author's own review. */
  ownAuthorLabel?: string;
  /** Label used for the review title when the review has none. */
  titleLabel: string;
  /** Label used for the date when the review has none. */
  dateLabel: string;
  /** Id of the signed-in user, so their own review reads as theirs. */
  currentUserId?: string | number | null;
  /** Display name of the signed-in user. */
  currentUserName?: string | null;
}): ProductReview {
  const {
    locale,
    authorLabel,
    ownAuthorLabel,
    titleLabel,
    dateLabel,
    currentUserId,
    currentUserName,
  } = options;
  const isOwn = Boolean(currentUserId) && String(review.userId) === String(currentUserId);

  return {
    id: review.id,
    author: isOwn
      ? currentUserName || ownAuthorLabel || authorLabel
      : review.user?.name || authorLabel,
    avatar: isOwn ? null : review.user?.avatar || null,
    rating: Number(review.rating) || 0,
    date: review.createdAt
      ? new Date(review.createdAt).toLocaleDateString(locale, {
          year: "numeric",
          month: "long",
          day: "numeric",
        })
      : dateLabel,
    title: review.title || titleLabel,
    content: review.body || "",
    helpful: review.helpfulCount ?? 0,
    verified: Boolean(review.orderItemId),
    userVoted: review.userVoted ?? false,
  };
}
