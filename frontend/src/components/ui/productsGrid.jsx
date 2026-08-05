import ProductsCard from "@/components/ui/productsCard";

export function ProductsGrid({ products, className, isWishlist }) {
	return (
		<div className={`grid ${className} gap-2.5 md:gap-5`}>
			{products.map((value) => (
				<ProductsCard
					key={value.id}
					cardDetails={value}
					isWishlist={isWishlist}
				/>
			))}
		</div>
	);
}
