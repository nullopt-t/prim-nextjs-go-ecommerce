interface PdpSkeletonProps {
  /** Screen-reader text announced while the skeleton is visible. */
  loadingLabel: string;
}

/**
 * Layout-matched skeleton for the product details page. Mirrors the loaded
 * page's geometry (breadcrumb, 7/5 grid, thumbnail rail, buy box) so client
 * navigations swap in place instead of collapsing to a spinner (CLS).
 */
export function PdpSkeleton({ loadingLabel }: PdpSkeletonProps) {
  return (
    <div role="status" aria-busy="true" className="mx-auto w-full max-w-7xl">
      <span className="sr-only">{loadingLabel}</span>

      {/* Breadcrumb */}
      <div className="mb-4 h-4 w-64 animate-pulse rounded bg-secondary/60 sm:mb-6" />

      {/* 7/5 media + buy box grid */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-12 lg:gap-8 xl:gap-10">
        {/* Gallery: vertical thumbnail rail + square main image */}
        <div className="relative lg:col-span-7">
          <div className="absolute inset-y-0 start-0 hidden w-20 flex-col gap-2.5 md:flex lg:w-24">
            {Array.from({ length: 4 }).map((_, index) => (
              <div
                key={index}
                className="aspect-square w-full shrink-0 animate-pulse rounded-xl bg-secondary/50"
              />
            ))}
          </div>
          <div className="md:ms-20 lg:ms-24">
            <div className="aspect-square animate-pulse rounded-2xl border border-border bg-secondary/50" />
          </div>
        </div>

        {/* Buy box card */}
        <div className="flex flex-col gap-5 lg:col-span-5">
          <div className="flex flex-col gap-5 rounded-2xl border border-border bg-card p-5 shadow-sm sm:p-7">
            <div className="flex flex-col gap-2">
              <div className="h-3 w-24 animate-pulse rounded bg-secondary/60" />
              <div className="h-8 w-11/12 animate-pulse rounded bg-secondary/70 sm:h-9 lg:h-10" />
            </div>

            <div className="mb-1 h-4 w-44 animate-pulse rounded bg-secondary/60" />

            {/* Price */}
            <div className="flex items-end gap-3 border-y border-border/70 py-4">
              <div className="h-10 w-36 animate-pulse rounded bg-secondary/70" />
              <div className="h-5 w-20 animate-pulse rounded bg-secondary/40" />
            </div>

            {/* Variant axis pills */}
            <div className="flex flex-col gap-2">
              <div className="h-3.5 w-20 animate-pulse rounded bg-secondary/60" />
              <div className="flex flex-wrap gap-2">
                {Array.from({ length: 4 }).map((_, index) => (
                  <div
                    key={index}
                    className="h-9 w-20 animate-pulse rounded-full bg-secondary/50"
                  />
                ))}
              </div>
            </div>
            <div className="flex flex-col gap-2">
              <div className="h-3.5 w-16 animate-pulse rounded bg-secondary/60" />
              <div className="flex flex-wrap gap-2">
                {Array.from({ length: 3 }).map((_, index) => (
                  <div
                    key={index}
                    className="h-9 w-16 animate-pulse rounded-full bg-secondary/50"
                  />
                ))}
              </div>
            </div>

            {/* Quantity + purchase actions */}
            <div className="h-12 w-full animate-pulse rounded-lg bg-secondary/50" />
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <div className="h-12 animate-pulse rounded-xl bg-secondary/50" />
              <div className="h-12 animate-pulse rounded-xl bg-secondary/50" />
            </div>

            {/* Trust pills */}
            <div className="flex flex-wrap gap-2">
              {Array.from({ length: 3 }).map((_, index) => (
                <div
                  key={index}
                  className="h-6 w-28 animate-pulse rounded-full bg-secondary/40"
                />
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
