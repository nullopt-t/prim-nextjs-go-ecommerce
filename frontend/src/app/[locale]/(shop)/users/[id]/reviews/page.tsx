import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { CircleUser } from "lucide-react";
import { getReviewsByUser } from "@/shared/lib/server/catalog";
import { mapReview } from "@/modules/catalog/utils/mapReview";
import { UserReviewsList } from "@/modules/catalog/components/ui/UserReviewsList";
import type { ProductReview, Review } from "@/modules/catalog/types";

type UserReviewsPageProps = {
  params: Promise<{ locale: string; id: string }>;
};

export async function generateMetadata({
  params,
}: UserReviewsPageProps): Promise<Metadata> {
  const { locale, id } = await params;
  const t = await getTranslations({ locale, namespace: "reviewerProfile" });
  const { reviews } = await getReviewsByUser(id);
  const name = reviews[0]?.user?.name || t("anonymous");
  return {
    title: t("title", { name }),
    description: t("description", { name }),
  };
}

/**
 * A reviewer's public profile: every approved review they wrote, each linked
 * to the product it was written about. The first page is server-rendered;
 * further pages stream in through the pager.
 */
export default async function UserReviewsPage({ params }: UserReviewsPageProps) {
  const { locale, id } = await params;
  const t = await getTranslations({ locale, namespace: "reviewerProfile" });
  const { reviews: rawReviews, nextCursor, hasMore } = await getReviewsByUser(id);

  const initialReviews: ProductReview[] = rawReviews.map((r: Review) =>
    mapReview(r, {
      locale,
      authorLabel: t("anonymous"),
      titleLabel: t("noTitle"),
      dateLabel: t("recently"),
    })
  );
  const authorName = rawReviews[0]?.user?.name || t("anonymous");

  return (
    <section className="mx-auto w-full max-w-6xl px-4 py-8 sm:px-6 sm:py-10">
      <div className="flex items-center gap-3 border-b border-border pb-5">
        <CircleUser className="size-11 shrink-0 text-muted-foreground" aria-hidden="true" />
        <h1 dir="auto" className="text-lg font-bold text-foreground sm:text-xl">
          {authorName}
        </h1>
      </div>
      <UserReviewsList
        userId={id}
        initialReviews={initialReviews}
        initialHasMore={hasMore}
        initialCursor={nextCursor}
      />
    </section>
  );
}
