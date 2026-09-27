import { ProductsCard, CardDetails } from "@/components/ui/productsCard";

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
    <div className={`grid ${className} gap-3 sm:gap-4 md:gap-5`}>
      {products.map((item) => (
        <ProductsCard
          key={item.id}
          cardDetails={item}
          isWishlist={isWishlist}
        />
      ))}
    </div>
  );
}

export default ProductsGrid;
