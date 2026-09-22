import { ProductsCard, CardDetails } from "@/components/ui/productsCard";
import { Link } from "@/i18n/navigation";

interface ProductsGridProps {
  products: (CardDetails & { slug?: string })[];
  className?: string;
  isWishlist?: boolean;
}

export function ProductsGrid({
  products,
  className = "",
  isWishlist = false,
}: ProductsGridProps) {
  return (
    <div className={`grid ${className} gap-2.5 md:gap-5`}>
      {products.map((item) => (
        <Link key={item.id} href={`/products/${item.slug || item.id}`} className="contents">
          <ProductsCard
            cardDetails={item}
            isWishlist={isWishlist}
          />
        </Link>
      ))}
    </div>
  );
}
