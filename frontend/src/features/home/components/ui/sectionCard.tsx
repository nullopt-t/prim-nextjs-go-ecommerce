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
      className="p-2.5 py-5 capitalize text-foreground hover:text-accent-brand hover:border-accent-brand bg-card font-medium flex items-center justify-center rounded-sm border-2 border-border hover:bg-accent transition-all text-center"
    >
      {category}
    </Link>
  );
}
