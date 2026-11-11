import type { ReactNode } from "react";
import { cn } from "@/shared/lib/cn";

interface ProductSectionProps {
  /** Section heading. */
  title: string;
  /** Small label above the heading (used by the reviews section). */
  eyebrow?: string;
  /** Optional leading icon. */
  icon?: ReactNode;
  /** Optional supporting line under the heading. */
  description?: string;
  /** Optional control rendered at the end of the heading row. */
  action?: ReactNode;
  children: ReactNode;
  /** Extra classes for the section wrapper. */
  className?: string;
  /** Extra classes for the content wrapper. */
  contentClassName?: string;
}

/**
 * Shared chrome for every product detail section: one card, one heading
 * scale, one content rhythm. Before this the page mixed bordered cards,
 * bare sections and borderless headings, so sections behaved differently
 * from one another.
 */
export function ProductSection({
  title,
  eyebrow,
  icon,
  description,
  action,
  children,
  className,
  contentClassName,
}: ProductSectionProps) {
  return (
    <section className={cn("rounded-2xl border border-border bg-card p-5 shadow-sm sm:p-6", className)}>
      <div className="mb-4 flex flex-wrap items-start justify-between gap-3 border-b border-border/70 pb-3">
        <div className="min-w-0">
          {eyebrow ? (
            <p className="mb-1 text-[11px] font-bold uppercase tracking-[0.16em] text-accent-brand">
              {eyebrow}
            </p>
          ) : null}
          <h2 className="flex items-center gap-2 text-base font-semibold text-foreground sm:text-lg">
            {icon}
            {title}
          </h2>
          {description ? (
            <p className="mt-1 text-xs leading-relaxed text-muted-foreground">{description}</p>
          ) : null}
        </div>
        {action ? <div className="shrink-0">{action}</div> : null}
      </div>
      <div className={contentClassName}>{children}</div>
    </section>
  );
}
