"use client";

import { useEffect, useState } from "react";
import { Link } from "@/i18n/navigation";
import { userService } from "@/services/user";
import { Stars } from "@/components/ui/stars";
import { Star, Trash2 } from "lucide-react";
import { toast } from "sonner";

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
const STATUS_BADGE: Record<Review["status"], { label: string; className: string }> = {
  pending: {
    label: "Pending",
    className:
      "text-amber-600 bg-amber-50 dark:bg-amber-950/30 dark:text-amber-400",
  },
  approved: {
    label: "Approved",
    className:
      "text-emerald-600 bg-emerald-50 dark:bg-emerald-950/30 dark:text-emerald-400",
  },
  rejected: {
    label: "Rejected",
    className: "text-red-600 bg-red-50 dark:bg-red-950/30 dark:text-red-400",
  },
};

// ── Skeleton card ──────────────────────────────────────────────────────────────
function ReviewSkeleton() {
  return <div className="animate-pulse bg-secondary rounded-lg h-32 w-full" />;
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
  const badge = STATUS_BADGE[review.status] ?? STATUS_BADGE.pending;
  const date = new Date(review.createdAt).toLocaleDateString();
  const bodyTruncated = review.body && review.body.length > 160 && !expanded;

  const handleDelete = async () => {
    setDeleting(true);
    try {
      await userService.deleteReview(review.id);
      toast.success("Review deleted.");
      onDelete(review.id);
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : "Failed to delete review.");
    } finally {
      setDeleting(false);
    }
  };

  return (
    <div className="bg-card border border-border rounded-lg shadow-sm p-4 flex flex-col gap-3">
      {/* Top row */}
      <div className="flex items-start justify-between gap-3">
        <div className="flex flex-col gap-1">
          <Stars starsNum={review.rating} />
          {review.title && (
            <p className="text-sm font-semibold text-foreground">{review.title}</p>
          )}
        </div>
        <div className="flex items-center gap-2 shrink-0">
          <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${badge.className}`}>
            {badge.label}
          </span>
          <button
            type="button"
            onClick={handleDelete}
            disabled={deleting}
            aria-label="Delete review"
            className="h-9 px-3 rounded-md text-sm font-medium text-destructive hover:bg-destructive/10 transition disabled:opacity-50"
          >
            <Trash2 className="size-4" />
          </button>
        </div>
      </div>

      {/* Body */}
      {review.body && (
        <div>
          <p
            className={`text-sm text-muted-foreground ${bodyTruncated ? "line-clamp-3" : ""}`}
          >
            {review.body}
          </p>
          {review.body.length > 160 && (
            <button
              type="button"
              onClick={() => setExpanded((v) => !v)}
              className="text-xs text-accent-brand hover:underline mt-1"
            >
              {expanded ? "Show less" : "Show more"}
            </button>
          )}
        </div>
      )}

      {/* Date */}
      <p className="text-xs text-muted-foreground">{date}</p>
    </div>
  );
}

// ── Main component ─────────────────────────────────────────────────────────────
export default function ReviewsContent() {
  const [reviews, setReviews] = useState<Review[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

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
        <h1 className="text-xl font-semibold text-foreground">My Reviews</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Manage your product reviews.
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
        <div className="bg-card border border-border rounded-lg p-12 flex flex-col items-center justify-center text-center gap-3 shadow-sm">
          <Star className="size-10 text-muted-foreground/50" />
          <p className="font-medium text-foreground">No reviews yet</p>
          <p className="text-sm text-muted-foreground">
            You haven&apos;t reviewed any products yet.
          </p>
          <Link
            href="/products"
            className="mt-2 h-9 px-4 inline-flex items-center rounded-md bg-accent-brand text-white text-sm font-medium hover:bg-accent-brand/90 transition shadow-sm"
          >
            Browse Products
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
