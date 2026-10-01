import { ProductsCard, CardDetails } from "@/components/ui/productsCard";

interface ProductsGridProps {
  products: (CardDetails & { slug?: string })[];
  className?: string;
  isWishlist?: boolean;
  onRemoveItem?: (item: CardDetails & { slug?: string }) => void;
}

export function ProductsGrid({
  products,
  className = "",
  isWishlist = false,
  onRemoveItem,
}: ProductsGridProps) {
  return (
    <div className={`grid ${className} gap-3 sm:gap-4 md:gap-5`}>
      {products.map((item) => (
        <ProductsCard
          key={item.wishlistItemId || (item.variantId ? `${item.id}-${item.variantId}` : item.id)}
          cardDetails={item}
          isWishlist={isWishlist}
          onRemove={onRemoveItem ? () => onRemoveItem(item) : undefined}
        />
      ))}
    </div>
  );
}

export default ProductsGrid;
