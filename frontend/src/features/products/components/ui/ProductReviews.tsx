import { useState } from "react";
import { User, ThumbsUp } from "lucide-react";
import { Stars } from "@/components/ui/stars";
import { toast } from "sonner";
import { ProductReview } from "../../domain/productDomain";

interface ProductReviewsProps {
  stars?: number;
  reviewsCount?: number;
  reviews: ProductReview[];
  onAddReview: (review: Omit<ProductReview, "id" | "date" | "helpful" | "verified">) => void;
  onVoteHelpful: (reviewId: string | number) => void;
}

export function ProductReviews({
  stars = 5,
  reviewsCount = 0,
  reviews,
  onAddReview,
  onVoteHelpful,
}: ProductReviewsProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [form, setForm] = useState({ rating: 0, title: "", content: "" });

  const handleSubmit = () => {
    if (!form.rating || !form.title.trim() || !form.content.trim()) return;
    onAddReview({
      author: "You",
      avatar: null,
      rating: form.rating,
      title: form.title.trim(),
      content: form.content.trim(),
    });
    setForm({ rating: 0, title: "", content: "" });
    setIsFormOpen(false);
    toast.success("Review submitted!");
  };

  const visibleReviews = isExpanded ? reviews : reviews.slice(0, 2);

  return (
    <section>
      <h2 className="text-2xl font-bold text-foreground mb-6">Customer Reviews</h2>
      <div className="bg-card border border-border rounded-2xl p-6 md:p-8 shadow-sm">
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-10">
          {/* Summary / Rating breakdown */}
          <div className="lg:col-span-4 flex flex-col">
            <div className="flex items-center gap-4 mb-4">
              <Stars starsNum={stars} />
              <span className="text-xl font-bold text-foreground">{stars} out of 5</span>
            </div>
            <p className="text-sm text-muted-foreground mb-6">{reviewsCount} global ratings</p>

            {/* Rating Bars */}
            <div className="flex flex-col gap-2 mb-8">
              {[
                { starVal: 5, pct: 72 },
                { starVal: 4, pct: 18 },
                { starVal: 3, pct: 6 },
                { starVal: 2, pct: 2 },
                { starVal: 1, pct: 2 },
              ].map((bar) => (
                <div key={bar.starVal} className="flex items-center gap-3 text-sm">
                  <span className="w-12 text-accent-brand hover:underline cursor-pointer font-medium whitespace-nowrap">
                    {bar.starVal} star
                  </span>
                  <div className="flex-1 h-4 bg-secondary/50 rounded-full overflow-hidden border border-border/50">
                    <div
                      className="h-full bg-amber-400 rounded-full"
                      style={{ width: `${bar.pct}%` }}
                    />
                  </div>
                  <span className="w-10 text-right text-muted-foreground">{bar.pct}%</span>
                </div>
              ))}
            </div>

            <div className="border-t border-border pt-6">
              <h4 className="font-bold text-foreground mb-2">Review this product</h4>
              <p className="text-sm text-muted-foreground mb-4">Share your thoughts with other customers</p>
              {!isFormOpen ? (
                <button
                  type="button"
                  onClick={() => setIsFormOpen(true)}
                  className="w-full py-2 px-4 rounded-lg bg-secondary text-foreground hover:bg-secondary/80 border border-border shadow-sm text-sm font-medium transition-colors cursor-pointer"
                >
                  Write a customer review
                </button>
              ) : (
                <div className="flex flex-col gap-4 mt-2 animate-in fade-in slide-in-from-top-2 duration-300">
                  <div className="flex flex-col gap-1.5">
                    <label className="text-sm font-semibold text-foreground">Rating</label>
                    <div className="flex items-center gap-1">
                      {[1, 2, 3, 4, 5].map((star) => (
                        <button
                          key={star}
                          type="button"
                          onClick={() => setForm({ ...form, rating: star })}
                          className={`text-2xl transition-colors hover:scale-110 cursor-pointer ${
                            form.rating >= star ? "text-amber-400" : "text-border"
                          }`}
                        >
                          ★
                        </button>
                      ))}
                    </div>
                  </div>

                  <div className="flex flex-col gap-1.5">
                    <label className="text-sm font-semibold text-foreground">Add a headline</label>
                    <input
                      type="text"
                      placeholder="What's most important to know?"
                      className="w-full px-3 py-2 text-sm rounded-lg border border-border bg-background focus:outline-none focus:ring-2 focus:ring-accent-brand focus:border-transparent transition-all"
                      value={form.title}
                      onChange={(e) => setForm({ ...form, title: e.target.value })}
                    />
                  </div>

                  <div className="flex flex-col gap-1.5">
                    <label className="text-sm font-semibold text-foreground">Add a written review</label>
                    <textarea
                      placeholder="What did you like or dislike? What did you use this product for?"
                      className="w-full px-3 py-2 text-sm rounded-lg border border-border bg-background min-h-[120px] resize-y focus:outline-none focus:ring-2 focus:ring-accent-brand focus:border-transparent transition-all"
                      value={form.content}
                      onChange={(e) => setForm({ ...form, content: e.target.value })}
                    />
                  </div>

                  <div className="flex gap-3 mt-2">
                    <button
                      type="button"
                      onClick={() => {
                        setIsFormOpen(false);
                        setForm({ rating: 0, title: "", content: "" });
                      }}
                      className="flex-1 py-2 px-4 rounded-lg bg-secondary text-foreground hover:bg-secondary/80 border border-border shadow-sm text-sm font-medium transition-colors cursor-pointer"
                    >
                      Cancel
                    </button>
                    <button
                      type="button"
                      onClick={handleSubmit}
                      className="flex-1 py-2 px-4 rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 shadow-sm text-sm font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
                      disabled={!form.rating || !form.title.trim() || !form.content.trim()}
                    >
                      Submit
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Reviews list */}
          <div className="lg:col-span-8 flex flex-col gap-8">
            {visibleReviews.map((review) => (
              <div key={review.id} className="flex flex-col border-b border-border/50 pb-8 last:border-0 last:pb-0">
                <div className="flex items-center gap-3 mb-2">
                  {review.avatar ? (
                    <img src={review.avatar} alt={review.author} className="w-10 h-10 rounded-full object-cover" />
                  ) : (
                    <div className="w-10 h-10 rounded-full bg-secondary flex items-center justify-center text-muted-foreground">
                      <User className="size-5" />
                    </div>
                  )}
                  <span className="font-medium text-foreground">{review.author}</span>
                </div>

                <div className="flex items-center gap-2 mb-1">
                  <Stars starsNum={review.rating} />
                  <span className="font-bold text-foreground text-sm">{review.title}</span>
                </div>

                <p className="text-xs text-muted-foreground mb-2">{review.date}</p>

                {review.verified && (
                  <p className="text-xs font-semibold text-accent-brand mb-3">Verified Purchase</p>
                )}

                <p className="text-sm text-foreground leading-relaxed mb-4">
                  {review.content}
                </p>

                <div className="flex items-center gap-4">
                  <button
                    type="button"
                    onClick={() => onVoteHelpful(review.id)}
                    className={`flex items-center gap-1.5 text-xs font-medium transition-colors border rounded-full px-3 py-1.5 cursor-pointer ${
                      review.userVoted
                        ? "border-accent-brand bg-accent-brand/10 text-accent-brand"
                        : "border-border text-muted-foreground hover:text-foreground hover:bg-secondary"
                    }`}
                  >
                    <ThumbsUp className={`size-3.5 ${review.userVoted ? "fill-current" : ""}`} />
                    Helpful
                  </button>
                  <span className="text-xs text-muted-foreground">
                    {review.helpful} people found this helpful
                  </span>
                </div>
              </div>
            ))}

            {reviews.length > 2 && (
              <button
                type="button"
                onClick={() => setIsExpanded(!isExpanded)}
                className="mt-4 w-full sm:w-auto self-center py-2 px-6 rounded-lg bg-secondary text-foreground hover:bg-secondary/80 border border-border shadow-sm text-sm font-medium transition-colors cursor-pointer"
              >
                {isExpanded ? "Show less" : "Show more"}
              </button>
            )}
          </div>
        </div>
      </div>
    </section>
  );
}
