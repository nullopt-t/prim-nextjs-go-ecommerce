import { Link } from "@/i18n/navigation";

interface SectionCardProps {
  category: string;
  slug?: string;
}

export default function SectionCard({ category, slug }: SectionCardProps) {
  const href = slug ? `/products?category=${slug}` : "/products";

  return (
    <Link
      href={href}
      className="p-3 py-5 capitalize text-foreground hover:text-accent-brand hover:border-accent-brand bg-card font-semibold text-sm flex items-center justify-center rounded-xl border border-border hover:bg-accent/40 shadow-2xs hover:shadow-xs transition-all text-center"
    >
      {category}
    </Link>
  );
}
